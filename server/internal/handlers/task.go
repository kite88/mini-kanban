package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mini-kanban/internal/db"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/models"
	"mini-kanban/internal/oplog"
	"mini-kanban/internal/storage"
	"mini-kanban/internal/utils"
)

// attachmentInput 前端回传的附件描述：文件已先通过 /api/uploads 上传拿到 url
type attachmentInput struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type taskInput struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	AssigneeID  *models.SnowflakeID `json:"assigneeId"` // 负责人（用户雪花 ID）；nil 表示不变更
	DueDate     *time.Time          `json:"dueDate"`
	Status      string              `json:"status"`
	Attachments []attachmentInput   `json:"attachments"` // nil 表示本次不改动附件
}

// attachmentMeta 校验通过的附件元信息（文件已落盘）。
// 任务附件与项目资料附件的表结构相同，校验逻辑共用，只是写库时落到各自的表。
type attachmentMeta struct {
	Name string
	URL  string
	Size int64
}

// validateAttachments 校验附件地址（必须是本站上传的文件），返回去重后的元信息。
func validateAttachments(list []attachmentInput) ([]attachmentMeta, error) {
	rows := make([]attachmentMeta, 0, len(list))
	seen := map[string]bool{}
	for _, a := range list {
		name, ok := storage.NameFromURL(a.URL)
		if !ok {
			return nil, errors.New("附件地址非法")
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, info, ok := storage.Path(name); !ok {
			return nil, errors.New("附件「" + CleanFileName(a.Name) + "」不存在")
		} else {
			rows = append(rows, attachmentMeta{
				Name: CleanFileName(a.Name),
				URL:  storage.URL(name),
				Size: info.Size(),
			})
		}
	}
	return rows, nil
}

// buildAttachments 校验附件并构造待写库的任务附件行。
func buildAttachments(taskID models.UUID, list []attachmentInput) ([]models.TaskAttachment, error) {
	metas, err := validateAttachments(list)
	if err != nil {
		return nil, err
	}
	rows := make([]models.TaskAttachment, 0, len(metas))
	for _, m := range metas {
		rows = append(rows, models.TaskAttachment{TaskID: taskID, Name: m.Name, URL: m.URL, Size: m.Size})
	}
	return rows, nil
}

// insertAttachments 写入附件行（新建任务用：先校验、建完任务再写）
func insertAttachments(taskID models.UUID, rows []models.TaskAttachment) error {
	if len(rows) == 0 {
		return nil
	}
	for i := range rows {
		rows[i].TaskID = taskID
	}
	return db.DB.Create(&rows).Error
}

// writeAttachments 覆盖式写入附件行（校验已由 buildAttachments 完成）。
// 只删表记录不删磁盘文件：同一文件可能还被富文本正文引用。
// 这里的删除是「编辑时替换旧附件」，不是用户可见的删除动作，因此用 Unscoped 物理删除：
// 走软删除的话每次保存任务都会留一批历史行，表会无意义地膨胀。
func writeAttachments(taskID models.UUID, rows []models.TaskAttachment) error {
	for i := range rows {
		rows[i].TaskID = taskID
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("task_id = ?", taskID).Delete(&models.TaskAttachment{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

// projectID 解析路径里的项目 ID（UUID v7）
func projectID(c *gin.Context) (models.UUID, bool) {
	return parseUUID(c, "id")
}

func ListTasks(c *gin.Context) {
	pid, good := projectID(c)
	if !good {
		return
	}
	// 项目级可见性：非成员（且非管理员）看不到项目下的任务
	if _, ok := loadAccessibleProject(c, pid); !ok {
		return
	}
	u := middleware.CurrentUser(c)
	q := db.DB.Preload("Assignee").Preload("Creator").Preload("Attachments").Where("project_id = ?", pid)
	// 任务可见性：管理员看全部；其他人看自己创建、被指派给自己的，
	// 以及自己作为「项目创建人」的项目下的全部任务。
	// 注意 OR 条件必须自己加括号，否则会与 project_id 按 AND 优先级组合错。
	if u.Role != models.RoleAdmin {
		owned := db.DB.Model(&models.Project{}).Select("id").Where("owner_id = ?", u.ID)
		q = q.Where("(creator_id = ? OR assignee_id = ? OR project_id IN (?))", u.ID, u.ID, owned)
	}
	var tasks []models.Task
	q.Order("position asc, id asc").Find(&tasks)
	ok(c, tasks)
}

func CreateTask(c *gin.Context) {
	pid, good := projectID(c)
	if !good {
		return
	}
	if _, ok := loadAccessibleProject(c, pid); !ok {
		return
	}
	// 仅凭「任务参与人」身份进入项目的人（被指派的非成员）不能新建任务
	if !isProjectMember(middleware.CurrentUser(c), pid) {
		fail(c, http.StatusForbidden, "只有项目成员可以创建任务")
		return
	}
	var req taskInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if req.Title == "" {
		badRequest(c, "标题不能为空")
		return
	}
	status := req.Status
	if !models.ValidStatus(status) {
		status = "todo"
	}
	// 放到目标列末尾
	var maxPos float64
	db.DB.Model(&models.Task{}).Where("project_id = ? AND status = ?", pid, status).
		Select("COALESCE(MAX(position), 0)").Scan(&maxPos)

	// 附件先校验（不落库）：主键在任务建完后的 insertAttachments 里补上
	attachRows, err := buildAttachments(models.UUID{}, req.Attachments)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	t := models.Task{
		ProjectID:   pid,
		Title:       req.Title,
		Description: utils.SanitizeRichText(req.Description),
		AssigneeID:  req.AssigneeID,
		CreatorID:   middleware.CurrentUser(c).ID,
		DueDate:     req.DueDate,
		Status:      status,
		Position:    maxPos + 1000,
	}
	if err := db.DB.Create(&t).Error; err != nil {
		fail(c, http.StatusInternalServerError, "创建失败")
		return
	}
	if err := insertAttachments(t.ID, attachRows); err != nil {
		fail(c, http.StatusInternalServerError, "保存附件失败")
		return
	}
	db.DB.Preload("Assignee").Preload("Creator").Preload("Attachments").First(&t, t.ID)
	invalidate("tasks", "projects")
	summary := "所属项目 " + oplog.ProjectTitle(pid) + "；状态 " + models.StatusLabel(t.Status)
	if t.AssigneeID != nil {
		summary += "；负责人 " + oplog.UserName(t.AssigneeID)
	}
	if len(attachRows) > 0 {
		summary += "；附件 " + strconv.Itoa(len(attachRows)) + " 个"
	}
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionCreate,
		TargetType: models.TargetTask,
		TargetID:   t.ID.String(),
		TargetName: t.Title,
		Summary:    summary,
	})
	created(c, t)
}

// canManageTask 任务权限：管理员、任务创建人、被指派人，或任务所属项目的创建人。
// 查看、编辑、拖拽改状态、删除都是这一套规则。
func canManageTask(u *models.User, t *models.Task) bool {
	if u == nil {
		return false
	}
	if u.Role == models.RoleAdmin || t.CreatorID == u.ID {
		return true
	}
	if t.AssigneeID != nil && *t.AssigneeID == u.ID {
		return true
	}
	return projectOwnerID(t.ProjectID) == u.ID
}

func GetTask(c *gin.Context) {
	id, good := parseUUID(c, "id")
	if !good {
		return
	}
	var t models.Task
	if err := db.DB.Preload("Assignee").Preload("Creator").Preload("Attachments").First(&t, id).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}
	// 可见范围与可编辑范围一致：管理员、任务创建人或被指派人
	if !canManageTask(middleware.CurrentUser(c), &t) {
		fail(c, http.StatusForbidden, "无权查看该任务")
		return
	}
	ok(c, t)
}

func UpdateTask(c *gin.Context) {
	id, good := parseUUID(c, "id")
	if !good {
		return
	}
	var t models.Task
	if err := db.DB.First(&t, id).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}
	// 权限：管理员、任务创建人或被指派人
	if !canManageTask(middleware.CurrentUser(c), &t) {
		fail(c, http.StatusForbidden, "只有任务创建人、被指派人、项目创建人或管理员可编辑")
		return
	}
	var req taskInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	// 附件先校验（req.Attachments 为 nil 表示本次不改动附件）
	var attachRows []models.TaskAttachment
	if req.Attachments != nil {
		rows, err := buildAttachments(t.ID, req.Attachments)
		if err != nil {
			badRequest(c, err.Error())
			return
		}
		attachRows = rows
	}

	oldStatus := t.Status
	// 保留修改前的内容，供操作日志对比（值拷贝，互不影响）
	before := t
	if req.Title != "" {
		t.Title = req.Title
	}
	t.Description = utils.SanitizeRichText(req.Description)
	if req.AssigneeID != nil {
		t.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		t.DueDate = req.DueDate
	}
	if req.Status != "" && models.ValidStatus(req.Status) && req.Status != oldStatus {
		t.Status = req.Status
		var maxPos float64
		db.DB.Model(&models.Task{}).Where("project_id = ? AND status = ?", t.ProjectID, t.Status).
			Select("COALESCE(MAX(position), 0)").Scan(&maxPos)
		t.Position = maxPos + 1000
	}
	if err := db.DB.Save(&t).Error; err != nil {
		fail(c, http.StatusInternalServerError, "保存失败")
		return
	}
	if req.Attachments != nil {
		if err := writeAttachments(t.ID, attachRows); err != nil {
			fail(c, http.StatusInternalServerError, "保存附件失败")
			return
		}
	}
	db.DB.Preload("Assignee").Preload("Creator").Preload("Attachments").First(&t, t.ID)
	invalidate("tasks", "projects")
	// 描述是富文本，正文不进摘要；附件只记数量
	summary := oplog.ChangesSummary(
		oplog.Change{Label: "标题", From: before.Title, To: t.Title},
		oplog.Change{
			Label: "状态",
			From:  models.StatusLabel(before.Status),
			To:    models.StatusLabel(t.Status),
		},
		oplog.Change{
			Label: "负责人",
			From:  oplog.UserName(before.AssigneeID),
			To:    oplog.UserName(t.AssigneeID),
		},
		oplog.Change{Label: "截止时间", From: before.DueDate, To: t.DueDate},
	)
	if before.Description != t.Description {
		summary = oplog.AppendSummary(summary, "描述已更新")
	}
	if req.Attachments != nil {
		summary = oplog.AppendSummary(summary, "附件已更新（"+strconv.Itoa(len(attachRows))+" 个）")
	}
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionUpdate,
		TargetType: models.TargetTask,
		TargetID:   t.ID.String(),
		TargetName: t.Title,
		Summary:    summary,
	})
	ok(c, t)
}

// MoveTask 看板拖拽：改变任务状态并插入到目标列的指定位置
func MoveTask(c *gin.Context) {
	id, good := parseUUID(c, "id")
	if !good {
		return
	}
	var req struct {
		Status string `json:"status"`
		Index  int    `json:"index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if !models.ValidStatus(req.Status) {
		badRequest(c, "状态非法")
		return
	}
	var t models.Task
	if err := db.DB.First(&t, id).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}
	// 权限：拖拽改状态属于编辑的一部分，规则与 UpdateTask 一致
	if !canManageTask(middleware.CurrentUser(c), &t) {
		fail(c, http.StatusForbidden, "只有任务创建人、被指派人、项目创建人或管理员可移动任务")
		return
	}

	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	oldStatus := t.Status
	t.Status = req.Status
	if err := tx.Model(&t).Select("Status").Updates(&t).Error; err != nil {
		tx.Rollback()
		fail(c, http.StatusInternalServerError, "移动失败")
		return
	}

	// 目标列重排（排除自身后按 index 插入）
	var others []models.Task
	tx.Where("project_id = ? AND status = ? AND id <> ?", t.ProjectID, req.Status, t.ID).
		Order("position asc, id asc").Find(&others)
	if req.Index < 0 {
		req.Index = 0
	}
	if req.Index > len(others) {
		req.Index = len(others)
	}
	ordered := make([]models.Task, 0, len(others)+1)
	ordered = append(ordered, others[:req.Index]...)
	ordered = append(ordered, t)
	ordered = append(ordered, others[req.Index:]...)
	for i := range ordered {
		pos := float64((i + 1) * 1000)
		if ordered[i].ID == t.ID {
			t.Position = pos
		}
		tx.Model(&models.Task{}).Where("id = ?", ordered[i].ID).Update("position", pos)
	}

	// 源列重排
	if oldStatus != req.Status {
		var src []models.Task
		tx.Where("project_id = ? AND status = ?", t.ProjectID, oldStatus).Order("position asc, id asc").Find(&src)
		for i := range src {
			tx.Model(&models.Task{}).Where("id = ?", src[i].ID).Update("position", float64((i+1)*1000))
		}
	}

	if err := tx.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "移动失败")
		return
	}
	db.DB.Preload("Assignee").Preload("Creator").Preload("Attachments").First(&t, t.ID)
	invalidate("tasks", "projects")
	// 同列内拖动只调整顺序、状态没变，不算业务变更，不记日志
	if oldStatus != req.Status {
		oplog.Record(middleware.CurrentUser(c), oplog.Entry{
			Action:     models.ActionMove,
			TargetType: models.TargetTask,
			TargetID:   t.ID.String(),
			TargetName: t.Title,
			Summary: oplog.ChangesSummary(oplog.Change{
				Label: "状态",
				From:  models.StatusLabel(oldStatus),
				To:    models.StatusLabel(req.Status),
			}),
		})
	}
	ok(c, t)
}

func DeleteTask(c *gin.Context) {
	id, good := parseUUID(c, "id")
	if !good {
		return
	}
	var t models.Task
	if err := db.DB.First(&t, id).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return
	}
	// 权限：管理员、任务创建人、被指派人、项目创建人（与查看 / 编辑 / 拖拽同一套）
	if !canManageTask(middleware.CurrentUser(c), &t) {
		fail(c, http.StatusForbidden, "只有管理员、任务创建人、被指派人或项目创建人可删除")
		return
	}
	// 软删除：任务与附件共用同一批次时间戳（磁盘文件本来就保留，见 writeAttachments），
	// 之后可在回收箱里按批次恢复
	if err := softDeleteTaskData(t.ID); err != nil {
		fail(c, http.StatusInternalServerError, "删除失败")
		return
	}
	invalidate("tasks", "projects")
	// 软删除的数据仍在库里，日志只记动作、不留快照（要留快照的是回收箱的「彻底删除」）
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionDelete,
		TargetType: models.TargetTask,
		TargetID:   t.ID.String(),
		TargetName: t.Title,
		Summary: oplog.AppendSummary(
			"所属项目 "+oplog.ProjectTitle(t.ProjectID),
			"软删除，可在回收箱恢复",
		),
	})
	okMessage(c, "已删除")
}
