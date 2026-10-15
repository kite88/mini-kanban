package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID          UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string      `gorm:"size:200;not null" json:"title"`
	Description string      `gorm:"type:text" json:"description"`
	OwnerID     SnowflakeID `gorm:"index;not null" json:"ownerId"`
	Owner       *User       `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	// Members 关联成员：只有成员（和管理员）能看到该项目
	Members   []User    `gorm:"many2many:project_members" json:"members,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// DeletedAt 软删除标记：删除项目只置位该列，数据保留在库中，
	// 之后所有查询都会自动带上 deleted_at IS NULL（要查已删数据用 Unscoped）。
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*Project) TableName() string { return "projects" }

// BeforeCreate 分配 UUID v7 主键
func (p *Project) BeforeCreate(*gorm.DB) error { return fillUUID(&p.ID) }
