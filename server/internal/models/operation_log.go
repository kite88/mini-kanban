package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// 操作日志的「动作」。字符串直出接口，前端按值渲染中文标签。
const (
	ActionCreate  = "create"  // 新建
	ActionUpdate  = "update"  // 修改
	ActionMove    = "move"    // 看板拖拽改状态（仅任务）
	ActionDelete  = "delete"  // 删除（项目 / 任务为软删除；用户 / 部门为物理删除）
	ActionRestore = "restore" // 从回收箱恢复
	ActionPurge   = "purge"   // 回收箱彻底删除（物理删除）
)

// 操作日志的「对象类型」（项目 / 任务 / 项目资料 / 用户 / 部门）
const (
	TargetProject    = "project"
	TargetTask       = "task"
	TargetDoc        = "doc" // 项目资料
	TargetUser       = "user"
	TargetDepartment = "department"
)

// JSONText 以 text 列保存 JSON 字符串（数据快照）。
//
// 不用 []byte：PostgreSQL 会把 []byte 当 bytea 写入、读出来是转义串；
// 实现 driver.Valuer / sql.Scanner 后进出都是原始 JSON 文本，
// 再由 MarshalJSON 还原成结构化对象，前端可直接展开查看。
type JSONText []byte

func (j JSONText) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONText) Scan(v interface{}) error {
	switch x := v.(type) {
	case nil:
		*j = nil
	case []byte:
		*j = append((*j)[:0], x...)
	case string:
		*j = append((*j)[:0], x...)
	default:
		return fmt.Errorf("无法解析 JSON 快照字段：%T", v)
	}
	return nil
}

func (j JSONText) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// OperationLog 操作日志：记录「谁在什么时候对哪个对象做了什么」，并附变更摘要。
//
// 只存操作人 id 与当时的用户名快照、不挂 User 关联：
// 一来不会因操作人被删除而连带删日志，二来列表查询不必 Preload。
// Snapshot 只在物理删除（删用户 / 删部门 / 回收箱彻底删除）时写入，
// 因为这类数据删除后无从追溯，必须整份留档。
type OperationLog struct {
	ID           UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	OperatorID   SnowflakeID `gorm:"index" json:"operatorId"`
	OperatorName string      `gorm:"size:64" json:"operatorName"`
	TargetType   string      `gorm:"size:16;index" json:"targetType"`
	// TargetID 目标 id：目标可能是用户（雪花数字）也可能是其他表（UUID），
	// 所以统一按字符串保存，避免一列混两种类型。
	TargetID   string    `gorm:"size:64;index" json:"targetId"`
	TargetName string    `gorm:"size:200" json:"targetName"`
	Action     string    `gorm:"size:16;index" json:"action"`
	Summary    string    `gorm:"size:500" json:"summary"`
	Snapshot   JSONText  `gorm:"type:text" json:"snapshot,omitempty"`
	CreatedAt  time.Time `gorm:"index" json:"createdAt"`
}

// TableName 固定表名，不依赖 gorm 的复数化推断
func (*OperationLog) TableName() string { return "operation_logs" }

// BeforeCreate 分配 UUID v7 主键
func (l *OperationLog) BeforeCreate(*gorm.DB) error { return fillUUID(&l.ID) }
