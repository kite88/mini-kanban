package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mini-kanban/internal/db"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/models"
	"mini-kanban/internal/oplog"
)

type projectInput struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	MemberIDs   *[]models.SnowflakeID `json:"memberIds"` // 关联成员（用户雪花 ID）；nil 表示不变更，空数组表示只保留创建人
}

// canManageProject 项目创建者或管理员可以编辑 / 删除项目
func canManageProject(u *models.User, p *models.Project) bool {
	if u == nil {
		return false
	}
	return u.Role == models.RoleAdmin || p.OwnerID == u.ID
}

// projectOwnerID 项目创建人 id（任务与项目资料共用的「项目创建人也算管理者」判据）。
// 用切片接 Pluck：雪花 ID 是自定义类型，标量目标不保证被 gorm 正确扫描
// （oplog/snapshot.go 里 Pluck 雪花列也是这个写法）。
func projectOwnerID(id models.UUID) models.SnowflakeID {
	var ownerIDs []models.SnowflakeID
	db.DB.Model(&models.Project{}).Where("id = ?", id).Pluck("owner_id", &ownerIDs)
	if len(ownerIDs) == 0 {
		return 0
	}
	return ownerIDs[0]
}

// projectMemberCount 项目成员数（项目创建人始终在关联表里，直接数关联表即可）
func projectMemberCount(projectID models.UUID) int {
	var n int64
	db.DB.Model(&models.ProjectMember{}).Where("project_id = ?", projectID).Count(&n)
	return int(n)
}

// projectMemberScope 「当前用户被关联的项目 id」子查询
func projectMemberScope(u *models.User) *gorm.DB {
	return db.DB.Model(&models.ProjectMember{}).Select("project_id").Where("user_id = ?", u.ID)
}

// projectTaskScope 「当前用户有任务参与（自己创建 / 指派给自己）的项目 id」子查询
func projectTaskScope(u *models.User) *gorm.DB {
	return db.DB.Model(&models.Task{}).Select("project_id").
		Where("creator_id = ? OR assignee_id = ?", u.ID, u.ID)
}

// isProjectMember 是否为项目成员（管理员等同全体；项目创建人始终算成员）
func isProjectMember(u *models.User, pid models.UUID) bool {
	if u == nil {
		return false
	}
	if u.Role == models.RoleAdmin {
		return true
	}
	var n int64
	db.DB.Model(&models.ProjectMember{}).Where("project_id = ? AND user_id = ?", pid, u.ID).Count(&n)
	if n > 0 {
		return true
	}
	db.DB.Model(&models.Project{}).Where("id = ? AND owner_id = ?", pid, u.ID).Count(&n)
	return n > 0
}

// canViewProject 项目可见性：管理员、项目创建者、已被关联的成员，
// 或在该项目下有任务参与关系的人（任务创建人 / 被指派人）。
// 最后一种覆盖「任务指派给了非项目成员」：他能进看板，但按任务可见性
// 规则只会看到与自己相关的任务，且不能新建任务（见 isProjectMember）。
func canViewProject(u *models.User, p *models.Project) bool {
	if u == nil {
		return false
	}
	if u.Role == models.RoleAdmin || p.OwnerID == u.ID {
		return true
	}
	var n int64
	db.DB.Model(&models.ProjectMember{}).Where("project_id = ? AND user_id = ?", p.ID, u.ID).Count(&n)
	if n > 0 {
		return true
	}
	db.DB.Model(&models.Task{}).
		Where("project_id = ? AND (creator_id = ? OR assignee_id = ?)", p.ID, u.ID, u.ID).Count(&n)
	return n > 0
}

// loadAccessibleProject 取项目并校验可见性，供项目下的任务接口复用
func loadAccessibleProject(c *gin.Context, pid models.UUID) (*models.Project, bool) {
	var p models.Project
	if err := db.DB.First(&p, pid).Error; err != nil {
		fail(c, http.StatusNotFound, "项目不存在")
		return nil, false
	}
	if !canViewProject(middleware.CurrentUser(c), &p) {
		fail(c, http.StatusForbidden, "无权访问该项目")
		return nil, false
	}
	return &p, true
}

// setProjectMembers 覆盖式设置项目成员：去重、只保留真实存在的用户，且创建人必在其中。
// ids 为 nil 时不改动成员。
func setProjectMembers(projectID models.UUID, ownerID models.SnowflakeID, ids []models.SnowflakeID, replace bool) error {
	if !replace {
		return nil
	}
	uniq := make([]models.SnowflakeID, 0, len(ids)+1)
	seen := map[models.SnowflakeID]bool{}
	for _, id := range append([]models.SnowflakeID{ownerID}, ids...) {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		uniq = append(uniq, id)
	}
	var valid []models.SnowflakeID
	if err := db.DB.Model(&models.User{}).Where("id IN ?", uniq).Pluck("id", &valid).Error; err != nil {
		return err
	}
	rows := make([]models.ProjectMember, 0, len(valid))
	for _, id := range valid {
		rows = append(rows, models.ProjectMember{ProjectID: projectID, UserID: id})
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", projectID).Delete(&models.ProjectMember{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

func ListProjects(c *gin.Context) {
	u := middleware.CurrentUser(c)

	q := db.DB.Preload("Owner").Preload("Members").Order("id desc")
	// 项目可见性：管理员看全部；其他人看被关联的项目，
	// 外加自己有任务参与的项目（与 canViewProject 保持一致）
	if u.Role != models.RoleAdmin {
		q = q.Where("(id IN (?) OR id IN (?))", projectMemberScope(u), projectTaskScope(u))
	}
	var list []models.Project
	q.Find(&list)

	// 附带任务数量统计：数量与看板可见范围保持一致（非管理员只统计与自己相关的任务）
	type item struct {
		models.Project
		TaskCount int64 `json:"taskCount"`
	}
	res := make([]item, 0, len(list))
	for _, p := range list {
		q := db.DB.Model(&models.Task{}).Where("project_id = ?", p.ID)
		if u.Role != models.RoleAdmin {
			q = q.Where("(creator_id = ? OR assignee_id = ?)", u.ID, u.ID)
		}
		var n int64
		q.Count(&n)
		res = append(res, item{Project: p, TaskCount: n})
	}
	ok(c, res)
}

func GetProject(c *gin.Context) {
	id, good := parseUUID(c, "id")
	if !good {
		return
	}
	var p models.Project
	if err := db.DB.Preload("Owner").Preload("Members").First(&p, id).Error; err != nil {
		fail(c, http.StatusNotFound, "项目不存在")
		return
	}
	if !canViewProject(middleware.CurrentUser(c), &p) {
		fail(c, http.StatusForbidden, "无权查看该项目")
		return
	}
	ok(c, p)
}

func CreateProject(c *gin.Context) {
	var req projectInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if req.Title == "" {
		badRequest(c, "标题不能为空")
		return
	}
	u := middleware.CurrentUser(c)
	p := models.Project{Title: req.Title, Description: req.Description, OwnerID: u.ID}
	if err := db.DB.Create(&p).Error; err != nil {
		fail(c, http.StatusInternalServerError, "创建失败")
		return
	}
	// 关联成员：创建人始终可见，其余按勾选写入（不传则只关联创建人）
	var ids []models.SnowflakeID
	if req.MemberIDs != nil {
		ids = *req.MemberIDs
	}
	if err := setProjectMembers(p.ID, u.ID, ids, true); err != nil {
		fail(c, http.StatusInternalServerError, "关联成员失败："+err.Error())
		return
	}
	db.DB.Preload("Owner").Preload("Members").First(&p, p.ID)
	invalidate("projects")
	oplog.Record(u, oplog.Entry{
		Action:     models.ActionCreate,
		TargetType: models.TargetProject,
		TargetID:   p.ID.String(),
		TargetName: p.Title,
		Summary:    "成员 " + strconv.Itoa(len(p.Members)) + " 人",
	})
	created(c, p)
}

func UpdateProject(c *gin.Context) {
	id, good := parseUUID(c, "id")
	if !good {
		return
	}
	var p models.Project
	if err := db.DB.First(&p, id).Error; err != nil {
		fail(c, http.StatusNotFound, "项目不存在")
		return
	}
	// 权限：项目创建者或管理员（与删除项目同一规则）
	if !canManageProject(middleware.CurrentUser(c), &p) {
		fail(c, http.StatusForbidden, "只有项目创建者或管理员可编辑")
		return
	}
	var req projectInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	// 保留修改前的内容与成员数，供操作日志对比（值拷贝，互不影响）
	before := p
	oldMemberCount := projectMemberCount(p.ID)
	if req.Title != "" {
		p.Title = req.Title
	}
	p.Description = req.Description
	if err := db.DB.Save(&p).Error; err != nil {
		fail(c, http.StatusInternalServerError, "保存失败")
		return
	}
	// 关联成员：传了 memberIds 才覆盖（创建人始终保留）
	if req.MemberIDs != nil {
		if err := setProjectMembers(p.ID, p.OwnerID, *req.MemberIDs, true); err != nil {
			fail(c, http.StatusInternalServerError, "保存成员失败："+err.Error())
			return
		}
	}
	db.DB.Preload("Owner").Preload("Members").First(&p, p.ID)
	invalidate("projects")
	// 描述是富文本，正文本身不进摘要，只提示「已更新」
	summary := oplog.ChangesSummary(
		oplog.Change{Label: "标题", From: before.Title, To: p.Title},
		oplog.Change{
			Label: "成员",
			From:  strconv.Itoa(oldMemberCount) + " 人",
			To:    strconv.Itoa(projectMemberCount(p.ID)) + " 人",
		},
	)
	if before.Description != p.Description {
		summary = oplog.AppendSummary(summary, "描述已更新")
	}
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionUpdate,
		TargetType: models.TargetProject,
		TargetID:   p.ID.String(),
		TargetName: p.Title,
		Summary:    summary,
	})
	ok(c, p)
}

func DeleteProject(c *gin.Context) {
	id, good := parseUUID(c, "id")
	if !good {
		return
	}
	var p models.Project
	if err := db.DB.First(&p, id).Error; err != nil {
		fail(c, http.StatusNotFound, "项目不存在")
		return
	}
	if !canManageProject(middleware.CurrentUser(c), &p) {
		fail(c, http.StatusForbidden, "只有项目创建者或管理员可删除")
		return
	}
	// 软删除：项目、项目资料、任务、附件都只置 deleted_at，数据保留在库中，默认查询自动过滤。
	// 它们共用同一个批次时间戳，回收箱恢复时按它精确还原（见 trash.go）。
	var taskCount, docCount int64
	db.DB.Model(&models.Task{}).Where("project_id = ?", p.ID).Count(&taskCount)
	db.DB.Model(&models.ProjectDoc{}).Where("project_id = ?", p.ID).Count(&docCount)
	if err := softDeleteProjectData(p.ID); err != nil {
		fail(c, http.StatusInternalServerError, "删除失败")
		return
	}
	// project_members 刻意保留：它是恢复项目后还原可见性的依据；
	// 项目本身已不可见（列表 / 详情都过滤 deleted_at），残留关联不会造成越权。
	invalidate("projects", "tasks", "docs")
	// 软删除的数据仍在库里，日志只记动作、不留快照（要留快照的是「彻底删除」）
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionDelete,
		TargetType: models.TargetProject,
		TargetID:   p.ID.String(),
		TargetName: p.Title,
		Summary: "软删除，连带 " + strconv.FormatInt(taskCount, 10) + " 个任务、" +
			strconv.FormatInt(docCount, 10) + " 条项目资料（可在回收箱恢复）",
	})
	okMessage(c, "已删除")
}
