package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mini-kanban/internal/db"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/models"
	"mini-kanban/internal/oplog"
	"mini-kanban/internal/utils"
)

// ============================ 项目资料 ============================
//
// 项目资料 = 挂在项目下的文档型内容（需求文档 / 会议纪要 / 设计稿 / 项目密钥 / 测试账密），
// 表单只有「标题 + 富文本正文 + 附件」，没有负责人 / 状态 / 到期时间。
//
// 权限：
//   - 查看：项目可见即可（管理员、项目成员、项目创建人，以及项目下有任务参与关系的人）
//   - 新建：项目成员（与新建任务同一规则）
//   - 编辑 / 删除：必须是项目成员，且是管理员、资料创建人或项目创建人
//     （人一旦离开项目，就失去对自己旧资料的处置权；管理员不受此限）

// docInput 项目资料的表单：附件已先通过 /api/uploads 上传拿到 url
type docInput struct {
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Attachments []attachmentInput `json:"attachments"` // nil 表示本次不改动附件
}

// parseDocID 解析路径里的资料 ID（UUID v7）
func parseDocID(c *gin.Context) (models.UUID, bool) {
	return parseUUID(c, "id")
}

// canManageDoc 项目资料权限：必须是项目成员（管理员也算成员），
// 且是管理员、资料创建人或资料所属项目的创建人。
// 「必须是成员」这一条用于收口：被移出项目的人不能再处置自己的旧资料。
func canManageDoc(u *models.User, d *models.ProjectDoc) bool {
	if u == nil || !isProjectMember(u, d.ProjectID) {
		return false
	}
	return u.Role == models.RoleAdmin || d.CreatorID == u.ID || projectOwnerID(d.ProjectID) == u.ID
}

// buildDocAttachments 校验附件并构造待写库的资料附件行
func buildDocAttachments(docID models.UUID, list []attachmentInput) ([]models.ProjectDocAttachment, error) {
	metas, err := validateAttachments(list)
	if err != nil {
		return nil, err
	}
	rows := make([]models.ProjectDocAttachment, 0, len(metas))
	for _, m := range metas {
		rows = append(rows, models.ProjectDocAttachment{DocID: docID, Name: m.Name, URL: m.URL, Size: m.Size})
	}
	return rows, nil
}

// insertDocAttachments 写入附件行（新建资料用：先校验、建完资料再写）
func insertDocAttachments(docID models.UUID, rows []models.ProjectDocAttachment) error {
	if len(rows) == 0 {
		return nil
	}
	for i := range rows {
		rows[i].DocID = docID
	}
	return db.DB.Create(&rows).Error
}

// writeDocAttachments 覆盖式写入附件行（校验已由 buildDocAttachments 完成）。
// 与任务附件一致：只删表记录不删磁盘文件（同一文件可能还被富文本正文引用），
// 且用 Unscoped 物理删除，避免每次保存都留一批历史行。
func writeDocAttachments(docID models.UUID, rows []models.ProjectDocAttachment) error {
	for i := range rows {
		rows[i].DocID = docID
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("doc_id = ?", docID).Delete(&models.ProjectDocAttachment{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

// ListProjectDocs 项目资料列表：最近更新的排前面
func ListProjectDocs(c *gin.Context) {
	pid, good := projectID(c)
	if !good {
		return
	}
	if _, ok := loadAccessibleProject(c, pid); !ok {
		return
	}
	var docs []models.ProjectDoc
	db.DB.Preload("Creator").Preload("Attachments").
		Where("project_id = ?", pid).Order("updated_at desc, id desc").Find(&docs)
	ok(c, docs)
}

// CreateProjectDoc 新建项目资料
func CreateProjectDoc(c *gin.Context) {
	pid, good := projectID(c)
	if !good {
		return
	}
	if _, ok := loadAccessibleProject(c, pid); !ok {
		return
	}
	// 与新建任务同一规则：只有项目成员能往项目里加资料
	if !isProjectMember(middleware.CurrentUser(c), pid) {
		fail(c, http.StatusForbidden, "只有项目成员可以新增项目资料")
		return
	}
	var req docInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		badRequest(c, "标题不能为空")
		return
	}
	// 附件先校验（不落库）：主键在资料建完后的 insertDocAttachments 里补上
	attachRows, err := buildDocAttachments(models.UUID{}, req.Attachments)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	d := models.ProjectDoc{
		ProjectID: pid,
		Title:     title,
		Content:   utils.SanitizeRichText(req.Content),
		CreatorID: middleware.CurrentUser(c).ID,
	}
	if err := db.DB.Create(&d).Error; err != nil {
		fail(c, http.StatusInternalServerError, "创建失败")
		return
	}
	if err := insertDocAttachments(d.ID, attachRows); err != nil {
		fail(c, http.StatusInternalServerError, "保存附件失败")
		return
	}
	db.DB.Preload("Creator").Preload("Attachments").First(&d, d.ID)
	invalidate("docs")
	summary := "所属项目 " + oplog.ProjectTitle(pid)
	if len(attachRows) > 0 {
		summary += "；附件 " + strconv.Itoa(len(attachRows)) + " 个"
	}
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionCreate,
		TargetType: models.TargetDoc,
		TargetID:   d.ID.String(),
		TargetName: d.Title,
		Summary:    summary,
	})
	created(c, d)
}

// UpdateProjectDoc 编辑项目资料
func UpdateProjectDoc(c *gin.Context) {
	id, good := parseDocID(c)
	if !good {
		return
	}
	var d models.ProjectDoc
	if err := db.DB.First(&d, id).Error; err != nil {
		fail(c, http.StatusNotFound, "资料不存在")
		return
	}
	if !canManageDoc(middleware.CurrentUser(c), &d) {
		fail(c, http.StatusForbidden, "只有项目成员中的资料创建人、项目创建人或管理员可编辑")
		return
	}
	var req docInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	// 附件先校验（req.Attachments 为 nil 表示本次不改动附件）
	var attachRows []models.ProjectDocAttachment
	if req.Attachments != nil {
		rows, err := buildDocAttachments(d.ID, req.Attachments)
		if err != nil {
			badRequest(c, err.Error())
			return
		}
		attachRows = rows
	}

	// 保留修改前的内容，供操作日志对比（值拷贝，互不影响）
	before := d
	if title := strings.TrimSpace(req.Title); title != "" {
		d.Title = title
	}
	d.Content = utils.SanitizeRichText(req.Content)
	if err := db.DB.Save(&d).Error; err != nil {
		fail(c, http.StatusInternalServerError, "保存失败")
		return
	}
	if req.Attachments != nil {
		if err := writeDocAttachments(d.ID, attachRows); err != nil {
			fail(c, http.StatusInternalServerError, "保存附件失败")
			return
		}
	}
	db.DB.Preload("Creator").Preload("Attachments").First(&d, d.ID)
	invalidate("docs")
	// 正文是富文本，正文本身不进摘要；附件只记数量
	summary := oplog.ChangesSummary(oplog.Change{Label: "标题", From: before.Title, To: d.Title})
	if before.Content != d.Content {
		summary = oplog.AppendSummary(summary, "内容已更新")
	}
	if req.Attachments != nil {
		summary = oplog.AppendSummary(summary, "附件已更新（"+strconv.Itoa(len(attachRows))+" 个）")
	}
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionUpdate,
		TargetType: models.TargetDoc,
		TargetID:   d.ID.String(),
		TargetName: d.Title,
		Summary:    summary,
	})
	ok(c, d)
}

// DeleteProjectDoc 删除项目资料：软删除，与它的附件共用同一批次时间戳
func DeleteProjectDoc(c *gin.Context) {
	id, good := parseDocID(c)
	if !good {
		return
	}
	var d models.ProjectDoc
	if err := db.DB.First(&d, id).Error; err != nil {
		fail(c, http.StatusNotFound, "资料不存在")
		return
	}
	if !canManageDoc(middleware.CurrentUser(c), &d) {
		fail(c, http.StatusForbidden, "只有项目成员中的资料创建人、项目创建人或管理员可删除")
		return
	}
	if err := softDeleteDocData(d.ID); err != nil {
		fail(c, http.StatusInternalServerError, "删除失败")
		return
	}
	invalidate("docs")
	// 软删除的数据仍在库里，日志只记动作、不留快照
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionDelete,
		TargetType: models.TargetDoc,
		TargetID:   d.ID.String(),
		TargetName: d.Title,
		Summary: oplog.AppendSummary(
			"所属项目 "+oplog.ProjectTitle(d.ProjectID),
			"软删除，可在回收箱恢复",
		),
	})
	okMessage(c, "已删除")
}
