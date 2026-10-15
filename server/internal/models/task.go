package models

import (
	"time"

	"gorm.io/gorm"
)

// TaskStatuses 任务状态，对应看板的各列（顺序即看板列顺序）
var TaskStatuses = []string{"todo", "doing", "done", "accepted"}

func ValidStatus(s string) bool {
	for _, v := range TaskStatuses {
		if v == s {
			return true
		}
	}
	return false
}

type Task struct {
	ID          UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID   UUID         `gorm:"type:uuid;index;not null" json:"projectId"`
	Title       string       `gorm:"size:200;not null" json:"title"`
	Description string       `gorm:"type:text" json:"description"`
	AssigneeID  *SnowflakeID `gorm:"index" json:"assigneeId"`
	Assignee    *User        `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	CreatorID   SnowflakeID  `gorm:"index" json:"creatorId"` // 任务创建人，删除权限判据之一
	Creator     *User        `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Status      string       `gorm:"size:16;default:'todo';index" json:"status"`
	Position    float64      `gorm:"default:0" json:"position"`
	DueDate     *time.Time   `json:"dueDate"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	// DeletedAt 软删除标记：任务（含随项目一起删除的任务）只置位该列，数据保留在库中
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`

	// Attachments 任务附件列表（独立表 task_attachments，按需 Preload）
	Attachments []TaskAttachment `gorm:"foreignKey:TaskID" json:"attachments"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*Task) TableName() string { return "tasks" }

// BeforeCreate 分配 UUID v7 主键
func (t *Task) BeforeCreate(*gorm.DB) error { return fillUUID(&t.ID) }
