package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role 用户角色
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type User struct {
	// ID 雪花 ID：64 位纯数字，由 NewSnowflakeID 生成（不走数据库自增）
	ID       SnowflakeID `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Email    string      `gorm:"size:128;uniqueIndex;not null" json:"email"`
	Username string      `gorm:"size:64;not null" json:"username"`
	// DepartmentID 所属部门，nil 表示未分配。
	// 只存外键、不挂 Department 关联对象：部门名由前端用部门列表按 id 映射，
	// 省掉每次查询用户的 Preload；归属合法性由 handlers 校验。
	DepartmentID *UUID     `gorm:"type:uuid;index" json:"departmentId"`
	Password     string    `gorm:"size:255;not null" json:"-"`
	Role         Role      `gorm:"size:16;default:'member'" json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*User) TableName() string { return "users" }

// BeforeCreate 为新用户分配雪花 ID
func (u *User) BeforeCreate(*gorm.DB) error {
	if u.ID == 0 {
		u.ID = NewSnowflakeID()
	}
	return nil
}

func (u *User) SetPassword(p string) error {
	h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(h)
	return nil
}

func (u *User) CheckPassword(p string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p)) == nil
}
