package models

import "time"

// ProjectMember 项目成员关联表（projects ↔ users 多对多）。
// 显式声明模型是为了让 migrate 能建表；列名与 gorm many2many 的默认约定一致。
// 复合主键：项目用 UUID、用户用雪花 ID（哪个类型跟哪张表，见 id.go 的主键约定）。
type ProjectMember struct {
	ProjectID UUID        `gorm:"type:uuid;primaryKey" json:"projectId"`
	UserID    SnowflakeID `gorm:"primaryKey" json:"userId"`
	CreatedAt time.Time   `json:"createdAt"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*ProjectMember) TableName() string { return "project_members" }
