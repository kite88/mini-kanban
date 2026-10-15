package models

import (
	"time"

	"gorm.io/gorm"
)

// UploadedFile 上传登记：文件落盘时同步登记一行，记录上传者。
//
// 用途是附件下载鉴权（见 handlers/canDownloadFile）：
//   - 被资料 / 任务引用（附件行或富文本正文）的文件，按该数据的可见性判定；
//   - 还没有被任何数据引用的临时文件（编辑器里刚粘贴的图片、上传后尚未保存的附件），
//     只允许上传者本人下载。没有这张表的话，这类文件无法判断归属，
//     任何登录用户拿到地址都能下。
type UploadedFile struct {
	ID         UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	Name       string      `gorm:"size:80;uniqueIndex;not null" json:"name"` // 落盘文件名（32 位随机 hex + 白名单扩展名）
	UploaderID SnowflakeID `gorm:"index" json:"uploaderId"`
	CreatedAt  time.Time   `json:"createdAt"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*UploadedFile) TableName() string { return "uploaded_files" }

// BeforeCreate 分配 UUID v7 主键
func (f *UploadedFile) BeforeCreate(*gorm.DB) error { return fillUUID(&f.ID) }
