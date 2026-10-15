package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mini-kanban/internal/models"
)

// 路径参数的 ID 分两类：
//   - 用户 id：雪花 ID，纯数字字符串 → parseUserID；
//   - 其他表的 id：UUID v7 → parseUUID。
//
// 解析失败统一返回 400（而不是 404），便于前端区分「参数写错」与「数据不存在」。

// parseUUID 解析路径里的 UUID 参数
func parseUUID(c *gin.Context, name string) (models.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param(name)))
	if err != nil {
		badRequest(c, "ID 非法")
		return models.UUID{}, false
	}
	return id, true
}

// parseUserID 解析用户 id（雪花 ID）
func parseUserID(c *gin.Context, name string) (models.SnowflakeID, bool) {
	id, err := models.ParseSnowflakeID(c.Param(name))
	if err != nil {
		badRequest(c, "用户 ID 非法")
		return 0, false
	}
	return id, true
}
