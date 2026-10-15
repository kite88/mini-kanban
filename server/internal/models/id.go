package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

// 主键约定：
//   - 用户表（users）：雪花 ID，64 位纯数字，由本进程生成（SnowflakeID）；
//   - 其他所有表：UUID v7，时间有序，由 GORM 的 BeforeCreate 钩子生成（见 fillUUID）；
//   - 指向 users 的外键列一律用 SnowflakeID，其余外键用 uuid.UUID。
//
// 前端 JavaScript 的安全整数只有 53 位，雪花 ID 会溢出，
// 所以 SnowflakeID 在 JSON 里统一按字符串收发（MarshalJSON / UnmarshalJSON）。

// UUID 是 uuid.UUID 的别名：业务代码里统一写 models.UUID，避免到处导入 uuid 包
type UUID = uuid.UUID

// SnowflakeID 用户表主键
type SnowflakeID int64

// snowflakeNode 进程内的雪花节点（单例），由 InitID 初始化
var snowflakeNode *snowflake.Node

// InitID 初始化雪花节点。nodeID 用于多实例部署时区分机器（单机固定 0 即可）。
// 必须在任何写库操作之前调用（main 里早于 db.Init）。
func InitID(nodeID int64) error {
	n, err := snowflake.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("初始化雪花节点失败：%w", err)
	}
	snowflakeNode = n
	return nil
}

// NewSnowflakeID 生成一个新的用户主键（雪花 ID）
func NewSnowflakeID() SnowflakeID {
	if snowflakeNode == nil {
		// 说明启动流程漏了 models.InitID，属于编码错误，直接暴露出来
		panic("models: 雪花节点未初始化（缺少 models.InitID 调用）")
	}
	return SnowflakeID(snowflakeNode.Generate().Int64())
}

// fillUUID 给空主键补一个 UUID v7
func fillUUID(id *uuid.UUID) error {
	if id == nil || *id != uuid.Nil {
		return nil
	}
	v, err := uuid.NewV7()
	if err != nil {
		return err
	}
	*id = v
	return nil
}

// String 便于日志与拼接
func (s SnowflakeID) String() string { return strconv.FormatInt(int64(s), 10) }

// IsZero 是否为空（未赋值）
func (s SnowflakeID) IsZero() bool { return s == 0 }

// MarshalJSON 按字符串输出：雪花 ID 有 19 位，超出 JS 安全整数范围
func (s SnowflakeID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(s), 10) + `"`), nil
}

// UnmarshalJSON 接受字符串或数字两种写法（前端一律传字符串）
func (s *SnowflakeID) UnmarshalJSON(b []byte) error {
	txt := strings.TrimSpace(string(b))
	if txt == "" || txt == "null" {
		*s = 0
		return nil
	}
	txt = strings.Trim(txt, `"`)
	if txt == "" || txt == "null" {
		*s = 0
		return nil
	}
	v, err := strconv.ParseInt(txt, 10, 64)
	if err != nil {
		return errors.New("用户 ID 必须是纯数字字符串")
	}
	*s = SnowflakeID(v)
	return nil
}

// Value 写库：雪花 ID 在 PostgreSQL 里就是 bigint
func (s SnowflakeID) Value() (driver.Value, error) { return int64(s), nil }

// Scan 读库：兼容 bigint 与字符串两种驱动返回值
func (s *SnowflakeID) Scan(v any) error {
	switch t := v.(type) {
	case nil:
		*s = 0
	case int64:
		*s = SnowflakeID(t)
	case int32:
		*s = SnowflakeID(t)
	case int:
		*s = SnowflakeID(t)
	case []byte:
		n, err := strconv.ParseInt(string(t), 10, 64)
		if err != nil {
			return err
		}
		*s = SnowflakeID(n)
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return err
		}
		*s = SnowflakeID(n)
	default:
		return fmt.Errorf("无法把 %T 解析为雪花 ID", v)
	}
	return nil
}

// ParseSnowflakeID 解析前端传来的用户 ID 字符串
func ParseSnowflakeID(s string) (SnowflakeID, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, errors.New("用户 ID 非法")
	}
	return SnowflakeID(v), nil
}

// 编译期断言：SnowflakeID 必须能直接作为 SQL 参数（GORM 依赖 driver.Valuer 读写主键）
var _ driver.Valuer = SnowflakeID(0)
