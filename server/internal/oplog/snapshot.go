package oplog

import (
	"mini-kanban/internal/db"
	"mini-kanban/internal/models"
)

// 数据快照：只在「物理删除」前采集，把即将从库里消失的内容整份留档。
//
// 结构与接口字段保持一致（复用模型的 json 标签），既方便人工排查，
// 也保留将来按快照重建数据的可能；外键只有 id 的字段额外补一份名称，
// 否则关联对象一删，快照里的 id 就看不出指向谁了。

// UserSnapshot 用户快照：账号信息 + 关联数据统计（不含密码哈希）
func UserSnapshot(id models.SnowflakeID) map[string]any {
	var u models.User
	if err := db.DB.First(&u, id).Error; err != nil {
		return nil
	}

	// 删除用户会解绑被指派任务、清理项目成员关联，这里把影响面一并留档
	var assignedTasks, createdTasks, ownedProjects, projectMembers int64
	db.DB.Unscoped().Model(&models.Task{}).Where("assignee_id = ?", id).Count(&assignedTasks)
	db.DB.Unscoped().Model(&models.Task{}).Where("creator_id = ?", id).Count(&createdTasks)
	db.DB.Unscoped().Model(&models.Project{}).Where("owner_id = ?", id).Count(&ownedProjects)
	db.DB.Model(&models.ProjectMember{}).Where("user_id = ?", id).Count(&projectMembers)

	return map[string]any{
		"user":           u, // 模型自带 json 标签，Password 已被 json:"-" 排除
		"departmentName": DepartmentName(u.DepartmentID),
		"related": map[string]any{
			"assignedTasks":  assignedTasks,
			"createdTasks":   createdTasks,
			"ownedProjects":  ownedProjects,
			"projectMembers": projectMembers,
		},
	}
}

// DepartmentSnapshot 部门快照：部门信息 + 剩余人数（正常情况下为 0，有成员时不允许删除）
func DepartmentSnapshot(id models.UUID) map[string]any {
	var d models.Department
	if err := db.DB.First(&d, id).Error; err != nil {
		return nil
	}
	var userCount int64
	db.DB.Model(&models.User{}).Where("department_id = ?", id).Count(&userCount)
	return map[string]any{
		"department": d,
		"userCount":  userCount,
	}
}

// ProjectSnapshot 项目快照：项目 + 成员 + 全部任务及其附件 + 全部项目资料及其附件。
// 彻底删除项目会连带清掉任务、资料、附件与成员关联，因此整棵树一起留档。
func ProjectSnapshot(id models.UUID) map[string]any {
	var p models.Project
	// Unscoped：彻底删除时项目已处于软删除状态
	if err := db.DB.Unscoped().Preload("Owner").First(&p, id).Error; err != nil {
		return nil
	}

	// 成员关联（项目删除时会被物理清理）
	var memberIDs []models.SnowflakeID
	db.DB.Model(&models.ProjectMember{}).Where("project_id = ?", id).
		Order("user_id asc").Pluck("user_id", &memberIDs)
	var members []models.User
	if len(memberIDs) > 0 {
		// 不带 Select：User 的 json:"-" 已经把密码排除在快照之外，整行取回时间字段更完整
		db.DB.Where("id IN ?", memberIDs).Order("id asc").Find(&members)
	}

	tasks := tasksWithAttachments("project_id = ?", id)
	docs := docsWithAttachments("project_id = ?", id)
	return map[string]any{
		"project":   p,
		"members":   members,
		"tasks":     tasks,
		"taskCount": len(tasks),
		"docs":      docs,
		"docCount":  len(docs),
	}
}

// TaskSnapshot 任务快照：任务 + 全部附件 + 所属项目标题
func TaskSnapshot(id models.UUID) map[string]any {
	tasks := tasksWithAttachments("id = ?", id)
	if len(tasks) == 0 {
		return nil
	}
	return map[string]any{
		"task":         tasks[0],
		"projectTitle": ProjectTitle(tasks[0].ProjectID),
	}
}

// DocSnapshot 项目资料快照：资料 + 全部附件 + 所属项目标题
func DocSnapshot(id models.UUID) map[string]any {
	docs := docsWithAttachments("id = ?", id)
	if len(docs) == 0 {
		return nil
	}
	return map[string]any{
		"doc":          docs[0],
		"projectTitle": ProjectTitle(docs[0].ProjectID),
	}
}

// docsWithAttachments 取项目资料（含已软删除的）并挂上全部附件。
// 与任务一样全程 Unscoped，且附件单独查再挂回去（Unscoped 不保证传播到预加载）。
func docsWithAttachments(where string, args ...any) []models.ProjectDoc {
	var docs []models.ProjectDoc
	db.DB.Unscoped().Preload("Creator").Where(where, args...).Order("id asc").Find(&docs)
	if len(docs) == 0 {
		return docs
	}

	ids := make([]models.UUID, 0, len(docs))
	for _, d := range docs {
		ids = append(ids, d.ID)
	}
	var attachments []models.ProjectDocAttachment
	db.DB.Unscoped().Where("doc_id IN ?", ids).Order("id asc").Find(&attachments)

	byDoc := make(map[models.UUID][]models.ProjectDocAttachment, len(ids))
	for _, a := range attachments {
		byDoc[a.DocID] = append(byDoc[a.DocID], a)
	}
	for i := range docs {
		// 没有附件时也给空数组，快照里的 attachments 才是 [] 而不是 null
		if rows := byDoc[docs[i].ID]; len(rows) > 0 {
			docs[i].Attachments = rows
		} else {
			docs[i].Attachments = []models.ProjectDocAttachment{}
		}
	}
	return docs
}

// tasksWithAttachments 取任务（含已软删除的）并挂上全部附件。
//
// 任务与附件都是软删除的，所以全程 Unscoped；附件单独查再挂回去，
// 而不是用 Preload("Attachments")：Unscoped 不保证传播到预加载，
// 漏掉附件就意味着快照不完整。
func tasksWithAttachments(where string, args ...any) []models.Task {
	var tasks []models.Task
	db.DB.Unscoped().Preload("Creator").Preload("Assignee").
		Where(where, args...).Order("id asc").Find(&tasks)
	if len(tasks) == 0 {
		return tasks
	}

	ids := make([]models.UUID, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.ID)
	}
	var attachments []models.TaskAttachment
	db.DB.Unscoped().Where("task_id IN ?", ids).Order("id asc").Find(&attachments)

	byTask := make(map[models.UUID][]models.TaskAttachment, len(ids))
	for _, a := range attachments {
		byTask[a.TaskID] = append(byTask[a.TaskID], a)
	}
	for i := range tasks {
		// 没有附件时也给空数组，快照里的 attachments 才是 [] 而不是 null
		if rows := byTask[tasks[i].ID]; len(rows) > 0 {
			tasks[i].Attachments = rows
		} else {
			tasks[i].Attachments = []models.TaskAttachment{}
		}
	}
	return tasks
}
