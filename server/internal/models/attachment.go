package models

import (
	"time"

	"gorm.io/gorm"
)

// TaskAttachment 任务附件：一个任务可有多个附件（富文本里内嵌的图片只落盘、不进这张表）。
// URL 存的是相对地址（/api/files/<随机名>），前端按 API 基址解析，便于换域名 / 端口。
type TaskAttachment struct {
	ID        UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TaskID    UUID      `gorm:"type:uuid;index;not null" json:"taskId"`
	Name      string    `gorm:"size:200;not null" json:"name"` // 上传时的原始文件名，仅用于展示与下载
	URL       string    `gorm:"size:255;not null" json:"url"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
	// DeletedAt 软删除标记：附件行随任务 / 项目删除置位保留，重启磁盘文件本身不受影响
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*TaskAttachment) TableName() string { return "task_attachments" }

// BeforeCreate 分配 UUID v7 主键
func (a *TaskAttachment) BeforeCreate(*gorm.DB) error { return fillUUID(&a.ID) }
