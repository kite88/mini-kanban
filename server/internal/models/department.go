package models

import (
	"time"

	"gorm.io/gorm"
)

// Department 部门：独立模块（增删改查），用户通过 department_id 关联
type Department struct {
	ID        UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Remark    string    `gorm:"size:200" json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*Department) TableName() string { return "departments" }

// BeforeCreate 分配 UUID v7 主键
func (d *Department) BeforeCreate(*gorm.DB) error { return fillUUID(&d.ID) }
