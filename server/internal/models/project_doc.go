package models

import (
	"time"

	"gorm.io/gorm"
)

// ProjectDoc 项目资料：挂在项目下的文档型内容（需求文档、会议纪要、设计稿、项目密钥、测试账密等）。
// 与任务的差别：没有负责人 / 状态 / 到期时间，只有标题 + 富文本正文 + 附件，
// 因此不参与看板与拖拽排序。
type ProjectDoc struct {
	ID        UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID UUID        `gorm:"type:uuid;index;not null" json:"projectId"`
	Title     string      `gorm:"size:200;not null" json:"title"`
	Content   string      `gorm:"type:text" json:"content"`
	CreatorID SnowflakeID `gorm:"index" json:"creatorId"` // 资料创建人，编辑 / 删除权限判据之一
	Creator   *User       `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
	// DeletedAt 软删除标记：删除资料只置位该列（可在回收箱恢复）；
	// 项目删除时与项目共用同一批次时间戳，项目恢复时按批次一并还原（见 handlers/trash.go）
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`

	// Attachments 资料附件列表（独立表 project_doc_attachments，按需 Preload）
	Attachments []ProjectDocAttachment `gorm:"foreignKey:DocID" json:"attachments"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*ProjectDoc) TableName() string { return "project_docs" }

// BeforeCreate 分配 UUID v7 主键
func (d *ProjectDoc) BeforeCreate(*gorm.DB) error { return fillUUID(&d.ID) }

// ProjectDocAttachment 项目资料附件，结构同 TaskAttachment。
// URL 存的是相对地址（/api/files/<随机名>），前端按 API 基址解析，便于换域名 / 端口。
type ProjectDocAttachment struct {
	ID        UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	DocID     UUID      `gorm:"type:uuid;index;not null" json:"docId"`
	Name      string    `gorm:"size:200;not null" json:"name"` // 上传时的原始文件名，仅用于展示与下载
	URL       string    `gorm:"size:255;not null" json:"url"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
	// DeletedAt 软删除标记：随资料 / 项目删除置位保留，磁盘文件本身不受影响
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*ProjectDocAttachment) TableName() string { return "project_doc_attachments" }

// BeforeCreate 分配 UUID v7 主键
func (a *ProjectDocAttachment) BeforeCreate(*gorm.DB) error { return fillUUID(&a.ID) }
