package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mini-kanban/internal/db"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/models"
	"mini-kanban/internal/oplog"
)

// ============================ 软删除「批次」 ============================
//
// 删除时把同一次操作涉及的记录（项目 / 任务 / 附件）用同一个时间戳写入 deleted_at，
// 恢复时按该时间戳就能精确圈出「同批被删」的记录：
// 恢复项目只会带回随它一起删掉的任务与附件，不会复活那些在项目删除之前
// 就被单独删掉（时间戳不同）的任务。
//
// 时间戳通过 UpdateColumn 而不是 gorm 的 Delete 写入，原因就是要让同批记录共享
// 同一个值；Delete 每条语句各取一次 now()，无法用于批次匹配。

// softDeleteProjectData 软删除项目及其资料、任务、附件（同一批次时间戳）
func softDeleteProjectData(projectID models.UUID) error {
	batch := time.Now()
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// 先附件、再主体、最后项目：附件靠「未删除的主体」子查询定位外键，
		// 顺序颠倒会让子查询变成空集而漏删附件。
		docIDs := tx.Model(&models.ProjectDoc{}).Select("id").Where("project_id = ?", projectID)
		if err := tx.Unscoped().Model(&models.ProjectDocAttachment{}).
			Where("doc_id IN (?)", docIDs).UpdateColumn("deleted_at", batch).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&models.ProjectDoc{}).
			Where("project_id = ? AND deleted_at IS NULL", projectID).
			UpdateColumn("deleted_at", batch).Error; err != nil {
			return err
		}
		taskIDs := tx.Model(&models.Task{}).Select("id").Where("project_id = ?", projectID)
		if err := tx.Unscoped().Model(&models.TaskAttachment{}).
			Where("task_id IN (?)", taskIDs).UpdateColumn("deleted_at", batch).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&models.Task{}).
			Where("project_id = ? AND deleted_at IS NULL", projectID).
			UpdateColumn("deleted_at", batch).Error; err != nil {
			return err
		}
		return tx.Unscoped().Model(&models.Project{}).
			Where("id = ? AND deleted_at IS NULL", projectID).
			UpdateColumn("deleted_at", batch).Error
	})
}

// softDeleteTaskData 软删除单个任务及其附件（同一批次时间戳）
func softDeleteTaskData(taskID models.UUID) error {
	batch := time.Now()
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&models.TaskAttachment{}).
			Where("task_id = ?", taskID).UpdateColumn("deleted_at", batch).Error; err != nil {
			return err
		}
		return tx.Unscoped().Model(&models.Task{}).
			Where("id = ? AND deleted_at IS NULL", taskID).
			UpdateColumn("deleted_at", batch).Error
	})
}

// softDeleteDocData 软删除单条项目资料及其附件（同一批次时间戳），可在回收箱恢复。
func softDeleteDocData(docID models.UUID) error {
	batch := time.Now()
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&models.ProjectDocAttachment{}).
			Where("doc_id = ?", docID).UpdateColumn("deleted_at", batch).Error; err != nil {
			return err
		}
		return tx.Unscoped().Model(&models.ProjectDoc{}).
			Where("id = ? AND deleted_at IS NULL", docID).
			UpdateColumn("deleted_at", batch).Error
	})
}

// restoreDocData 按批次恢复单条资料及其附件
func restoreDocData(d models.ProjectDoc) error {
	batch := d.DeletedAt
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&models.ProjectDocAttachment{}).
			Where("doc_id = ? AND deleted_at = ?", d.ID, batch).
			UpdateColumn("deleted_at", nil).Error; err != nil {
			return err
		}
		return tx.Unscoped().Model(&models.ProjectDoc{}).
			Where("id = ? AND deleted_at = ?", d.ID, batch).
			UpdateColumn("deleted_at", nil).Error
	})
}

// restoreProjectData 按批次恢复项目：只还原 deleted_at 与项目相同的资料 / 任务 / 附件
func restoreProjectData(p models.Project) error {
	batch := p.DeletedAt
	return db.DB.Transaction(func(tx *gorm.DB) error {
		docIDs := tx.Unscoped().Model(&models.ProjectDoc{}).Select("id").Where("project_id = ?", p.ID)
		if err := tx.Unscoped().Model(&models.ProjectDocAttachment{}).
			Where("doc_id IN (?) AND deleted_at = ?", docIDs, batch).
			UpdateColumn("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&models.ProjectDoc{}).
			Where("project_id = ? AND deleted_at = ?", p.ID, batch).
			UpdateColumn("deleted_at", nil).Error; err != nil {
			return err
		}
		taskIDs := tx.Unscoped().Model(&models.Task{}).Select("id").Where("project_id = ?", p.ID)
		if err := tx.Unscoped().Model(&models.TaskAttachment{}).
			Where("task_id IN (?) AND deleted_at = ?", taskIDs, batch).
			UpdateColumn("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&models.Task{}).
			Where("project_id = ? AND deleted_at = ?", p.ID, batch).
			UpdateColumn("deleted_at", nil).Error; err != nil {
			return err
		}
		return tx.Unscoped().Model(&models.Project{}).
			Where("id = ? AND deleted_at = ?", p.ID, batch).
			UpdateColumn("deleted_at", nil).Error
	})
}

// restoreTaskData 按批次恢复任务及其附件
func restoreTaskData(t models.Task) error {
	batch := t.DeletedAt
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&models.TaskAttachment{}).
			Where("task_id = ? AND deleted_at = ?", t.ID, batch).
			UpdateColumn("deleted_at", nil).Error; err != nil {
			return err
		}
		return tx.Unscoped().Model(&models.Task{}).
			Where("id = ? AND deleted_at = ?", t.ID, batch).
			UpdateColumn("deleted_at", nil).Error
	})
}

// purgeProjectData 彻底删除项目：连带其下全部资料、任务、附件与成员关联一并物理删除。
// 只做删除，事务由调用方开启——物理删除必须与它的数据快照写在同一个事务里。
func purgeProjectData(tx *gorm.DB, projectID models.UUID) error {
	// 顺序受外键约束限制（project_doc_attachments.doc_id → project_docs.id）
	docIDs := tx.Unscoped().Model(&models.ProjectDoc{}).Select("id").Where("project_id = ?", projectID)
	if err := tx.Unscoped().Where("doc_id IN (?)", docIDs).Delete(&models.ProjectDocAttachment{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&models.ProjectDoc{}).Error; err != nil {
		return err
	}
	// task_attachments.task_id → tasks.id
	taskIDs := tx.Unscoped().Model(&models.Task{}).Select("id").Where("project_id = ?", projectID)
	if err := tx.Unscoped().Where("task_id IN (?)", taskIDs).Delete(&models.TaskAttachment{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&models.Task{}).Error; err != nil {
		return err
	}
	if err := tx.Where("project_id = ?", projectID).Delete(&models.ProjectMember{}).Error; err != nil {
		return err
	}
	return tx.Unscoped().Delete(&models.Project{}, projectID).Error
}

// purgeTaskData 彻底删除单个任务及其附件（事务由调用方开启，原因同上）
func purgeTaskData(tx *gorm.DB, taskID models.UUID) error {
	if err := tx.Unscoped().Where("task_id = ?", taskID).Delete(&models.TaskAttachment{}).Error; err != nil {
		return err
	}
	return tx.Unscoped().Delete(&models.Task{}, taskID).Error
}

// purgeDocData 彻底删除单条项目资料及其附件（事务由调用方开启，原因同上）
func purgeDocData(tx *gorm.DB, docID models.UUID) error {
	if err := tx.Unscoped().Where("doc_id = ?", docID).Delete(&models.ProjectDocAttachment{}).Error; err != nil {
		return err
	}
	return tx.Unscoped().Delete(&models.ProjectDoc{}, docID).Error
}

// ============================ 回收箱接口 ============================

// trashProjectItem 回收箱里的项目条目：附带「随它一起被删的任务数 / 资料数」，
// 也就是恢复该项目时会一起回来的数量。
type trashProjectItem struct {
	models.Project
	TaskCount int64 `json:"taskCount"`
	DocCount  int64 `json:"docCount"`
}

// trashTaskItem 回收箱里的任务条目：所属项目标题用于展示（任务表里只有 projectId）
type trashTaskItem struct {
	models.Task
	ProjectTitle string `json:"projectTitle"`
}

// trashDocItem 回收箱里的项目资料条目：所属项目标题用于展示（资料表里只有 projectId）
type trashDocItem struct {
	models.ProjectDoc
	ProjectTitle string `json:"projectTitle"`
}

// ListTrash 回收箱列表：已删除的项目 + 已删除但所属项目仍健在的任务与项目资料。
//
// 每个用户只看到自己有权限处置的那些数据（与删除权限对称）：
// 项目按创建人、任务按「管理员 / 任务创建人 / 被指派人 / 项目创建人」、
// 资料按「管理员 / 资料创建人 / 项目创建人」且必须仍是项目成员。
// 随项目一起被删的任务与资料不重复出现，而是算进项目条目的 taskCount / docCount。
func ListTrash(c *gin.Context) {
	u := middleware.CurrentUser(c)

	pq := db.DB.Unscoped().Preload("Owner").Where("projects.deleted_at IS NOT NULL")
	if u.Role != models.RoleAdmin {
		pq = pq.Where("owner_id = ?", u.ID)
	}
	var projects []models.Project
	pq.Order("projects.deleted_at desc").Find(&projects)

	items := make([]trashProjectItem, 0, len(projects))
	for _, p := range projects {
		// 同批次（deleted_at 相同）的任务数 / 资料数 = 恢复时会一起回来的数量
		var n, dn int64
		db.DB.Unscoped().Model(&models.Task{}).
			Where("project_id = ? AND deleted_at = ?", p.ID, p.DeletedAt).
			Count(&n)
		db.DB.Unscoped().Model(&models.ProjectDoc{}).
			Where("project_id = ? AND deleted_at = ?", p.ID, p.DeletedAt).
			Count(&dn)
		items = append(items, trashProjectItem{Project: p, TaskCount: n, DocCount: dn})
	}

	// 所属项目仍在（未删除）的已删任务；项目已删的交给项目条目处理
	tq := db.DB.Unscoped().Preload("Creator").Preload("Assignee").
		Where("tasks.deleted_at IS NOT NULL").
		Where("project_id IN (?)", db.DB.Model(&models.Project{}).Select("id"))
	if u.Role != models.RoleAdmin {
		owned := db.DB.Model(&models.Project{}).Select("id").Where("owner_id = ?", u.ID)
		tq = tq.Where("(creator_id = ? OR assignee_id = ? OR project_id IN (?))", u.ID, u.ID, owned)
	}
	var tasks []models.Task
	tq.Order("tasks.deleted_at desc").Find(&tasks)

	// 已删除但所属项目仍健在的资料；项目已删的交给项目条目处理
	dq := db.DB.Unscoped().Preload("Creator").Preload("Attachments").
		Where("project_docs.deleted_at IS NOT NULL").
		Where("project_id IN (?)", db.DB.Model(&models.Project{}).Select("id"))
	if u.Role != models.RoleAdmin {
		// 处置权限与删除资料对称：必须是项目成员，且是资料创建人或项目创建人
		member := db.DB.Model(&models.ProjectMember{}).Select("project_id").Where("user_id = ?", u.ID)
		owned := db.DB.Model(&models.Project{}).Select("id").Where("owner_id = ?", u.ID)
		dq = dq.Where("project_id IN (?) AND (creator_id = ? OR project_id IN (?))", member, u.ID, owned)
	}
	var docs []models.ProjectDoc
	dq.Order("project_docs.deleted_at desc").Find(&docs)

	// 项目标题一次查出来，避免逐条查库（任务与资料共用）
	titles := map[models.UUID]string{}
	ids := make([]models.UUID, 0, len(tasks)+len(docs))
	for _, t := range tasks {
		ids = append(ids, t.ProjectID)
	}
	for _, d := range docs {
		ids = append(ids, d.ProjectID)
	}
	if len(ids) > 0 {
		var rows []models.Project
		db.DB.Select("id", "title").Where("id IN ?", ids).Find(&rows)
		for _, p := range rows {
			titles[p.ID] = p.Title
		}
	}
	taskItems := make([]trashTaskItem, 0, len(tasks))
	for _, t := range tasks {
		taskItems = append(taskItems, trashTaskItem{Task: t, ProjectTitle: titles[t.ProjectID]})
	}
	docItems := make([]trashDocItem, 0, len(docs))
	for _, d := range docs {
		docItems = append(docItems, trashDocItem{ProjectDoc: d, ProjectTitle: titles[d.ProjectID]})
	}

	ok(c, gin.H{"projects": items, "tasks": taskItems, "docs": docItems})
}

// loadDeletedProject 取回收箱里的项目（必须是已删除状态，否则说明客户端数据过期）
func loadDeletedProject(c *gin.Context) (*models.Project, bool) {
	id, good := parseUUID(c, "id")
	if !good {
		return nil, false
	}
	var p models.Project
	if err := db.DB.Unscoped().First(&p, id).Error; err != nil {
		fail(c, http.StatusNotFound, "项目不存在")
		return nil, false
	}
	if !p.DeletedAt.Valid {
		badRequest(c, "该项目未被删除")
		return nil, false
	}
	return &p, true
}

func loadDeletedTask(c *gin.Context) (*models.Task, bool) {
	id, good := parseUUID(c, "id")
	if !good {
		return nil, false
	}
	var t models.Task
	if err := db.DB.Unscoped().First(&t, id).Error; err != nil {
		fail(c, http.StatusNotFound, "任务不存在")
		return nil, false
	}
	if !t.DeletedAt.Valid {
		badRequest(c, "该任务未被删除")
		return nil, false
	}
	return &t, true
}

// loadDeletedDoc 取回收箱里的项目资料（必须是已删除状态，否则说明客户端数据过期）
func loadDeletedDoc(c *gin.Context) (*models.ProjectDoc, bool) {
	id, good := parseUUID(c, "id")
	if !good {
		return nil, false
	}
	var d models.ProjectDoc
	if err := db.DB.Unscoped().First(&d, id).Error; err != nil {
		fail(c, http.StatusNotFound, "资料不存在")
		return nil, false
	}
	if !d.DeletedAt.Valid {
		badRequest(c, "该资料未被删除")
		return nil, false
	}
	return &d, true
}

// RestoreProject 恢复项目：权限与删除项目一致（创建人或管理员）
func RestoreProject(c *gin.Context) {
	p, good := loadDeletedProject(c)
	if !good {
		return
	}
	if !canManageProject(middleware.CurrentUser(c), p) {
		fail(c, http.StatusForbidden, "只有项目创建者或管理员可恢复")
		return
	}
	if err := restoreProjectData(*p); err != nil {
		fail(c, http.StatusInternalServerError, "恢复失败："+err.Error())
		return
	}
	invalidate("projects", "tasks", "docs")
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionRestore,
		TargetType: models.TargetProject,
		TargetID:   p.ID.String(),
		TargetName: p.Title,
		Summary:    "从回收箱恢复，同批删除的任务、项目资料与附件一并还原",
	})
	okMessage(c, "已恢复")
}

// PurgeProject 彻底删除项目：物理删除，不可再恢复
func PurgeProject(c *gin.Context) {
	p, good := loadDeletedProject(c)
	if !good {
		return
	}
	op := middleware.CurrentUser(c)
	if !canManageProject(op, p) {
		fail(c, http.StatusForbidden, "只有项目创建者或管理员可彻底删除")
		return
	}
	// 物理删除前采集快照：项目、任务、附件、成员关联都会被清掉，快照是唯一的留档
	snapshot := oplog.ProjectSnapshot(p.ID)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := purgeProjectData(tx, p.ID); err != nil {
			return err
		}
		return oplog.RecordTx(tx, op, oplog.Entry{
			Action:     models.ActionPurge,
			TargetType: models.TargetProject,
			TargetID:   p.ID.String(),
			TargetName: p.Title,
			Summary:    "彻底删除，不可恢复；连带其下任务、项目资料、附件与成员关联一并物理删除",
			Snapshot:   snapshot,
		})
	})
	if err != nil {
		fail(c, http.StatusInternalServerError, "彻底删除失败："+err.Error())
		return
	}
	invalidate("projects", "tasks", "docs")
	okMessage(c, "已彻底删除")
}

// RestoreTask 恢复任务：权限与删除任务一致
func RestoreTask(c *gin.Context) {
	t, good := loadDeletedTask(c)
	if !good {
		return
	}
	if !canManageTask(middleware.CurrentUser(c), t) {
		fail(c, http.StatusForbidden, "只有管理员、任务创建人、被指派人或项目创建人可恢复")
		return
	}
	// 所属项目也已被删时，任务恢复出来在看板里依旧看不到，先恢复项目更直观
	var n int64
	db.DB.Unscoped().Model(&models.Project{}).
		Where("id = ? AND deleted_at IS NOT NULL", t.ProjectID).Count(&n)
	if n > 0 {
		badRequest(c, "任务所属项目已被删除，请先在回收箱恢复该项目")
		return
	}
	if err := restoreTaskData(*t); err != nil {
		fail(c, http.StatusInternalServerError, "恢复失败："+err.Error())
		return
	}
	invalidate("tasks", "projects")
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionRestore,
		TargetType: models.TargetTask,
		TargetID:   t.ID.String(),
		TargetName: t.Title,
		Summary: oplog.AppendSummary(
			"从回收箱恢复；所属项目 "+oplog.ProjectTitle(t.ProjectID),
			"同批删除的附件一并还原",
		),
	})
	okMessage(c, "已恢复")
}

// PurgeTask 彻底删除任务：物理删除，不可再恢复
func PurgeTask(c *gin.Context) {
	t, good := loadDeletedTask(c)
	if !good {
		return
	}
	op := middleware.CurrentUser(c)
	if !canManageTask(op, t) {
		fail(c, http.StatusForbidden, "只有管理员、任务创建人、被指派人或项目创建人可彻底删除")
		return
	}
	// 物理删除前采集快照：任务与附件都会被清掉，快照是唯一的留档
	snapshot := oplog.TaskSnapshot(t.ID)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := purgeTaskData(tx, t.ID); err != nil {
			return err
		}
		return oplog.RecordTx(tx, op, oplog.Entry{
			Action:     models.ActionPurge,
			TargetType: models.TargetTask,
			TargetID:   t.ID.String(),
			TargetName: t.Title,
			Summary: oplog.AppendSummary(
				"彻底删除，不可恢复；所属项目 "+oplog.ProjectTitle(t.ProjectID),
				"附件一并物理删除",
			),
			Snapshot: snapshot,
		})
	})
	if err != nil {
		fail(c, http.StatusInternalServerError, "彻底删除失败："+err.Error())
		return
	}
	invalidate("tasks", "projects")
	okMessage(c, "已彻底删除")
}

// RestoreDoc 恢复项目资料：权限与删除资料一致
// （必须是项目成员，且是资料创建人、项目创建人或管理员）
func RestoreDoc(c *gin.Context) {
	d, good := loadDeletedDoc(c)
	if !good {
		return
	}
	if !canManageDoc(middleware.CurrentUser(c), d) {
		fail(c, http.StatusForbidden, "只有项目成员中的资料创建人、项目创建人或管理员可恢复")
		return
	}
	// 所属项目也已被删时，资料恢复出来在看板里依旧看不到，先恢复项目更直观
	var n int64
	db.DB.Unscoped().Model(&models.Project{}).
		Where("id = ? AND deleted_at IS NOT NULL", d.ProjectID).Count(&n)
	if n > 0 {
		badRequest(c, "资料所属项目已被删除，请先在回收箱恢复该项目")
		return
	}
	if err := restoreDocData(*d); err != nil {
		fail(c, http.StatusInternalServerError, "恢复失败："+err.Error())
		return
	}
	invalidate("docs")
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionRestore,
		TargetType: models.TargetDoc,
		TargetID:   d.ID.String(),
		TargetName: d.Title,
		Summary: oplog.AppendSummary(
			"从回收箱恢复；所属项目 "+oplog.ProjectTitle(d.ProjectID),
			"同批删除的附件一并还原",
		),
	})
	okMessage(c, "已恢复")
}

// PurgeDoc 彻底删除项目资料：物理删除，不可再恢复
func PurgeDoc(c *gin.Context) {
	d, good := loadDeletedDoc(c)
	if !good {
		return
	}
	op := middleware.CurrentUser(c)
	if !canManageDoc(op, d) {
		fail(c, http.StatusForbidden, "只有项目成员中的资料创建人、项目创建人或管理员可彻底删除")
		return
	}
	// 物理删除前采集快照：资料与附件都会被清掉，快照是唯一的留档
	snapshot := oplog.DocSnapshot(d.ID)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := purgeDocData(tx, d.ID); err != nil {
			return err
		}
		return oplog.RecordTx(tx, op, oplog.Entry{
			Action:     models.ActionPurge,
			TargetType: models.TargetDoc,
			TargetID:   d.ID.String(),
			TargetName: d.Title,
			Summary: oplog.AppendSummary(
				"彻底删除，不可恢复；所属项目 "+oplog.ProjectTitle(d.ProjectID),
				"附件一并物理删除",
			),
			Snapshot: snapshot,
		})
	})
	if err != nil {
		fail(c, http.StatusInternalServerError, "彻底删除失败："+err.Error())
		return
	}
	invalidate("docs")
	okMessage(c, "已彻底删除")
}
