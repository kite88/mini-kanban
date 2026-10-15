package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-kanban/internal/cache"
	"mini-kanban/internal/i18n"
)

func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// okMessage 操作成功的文字提示（已删除 / 已恢复 / 已登出 等），按请求语言返回
func okMessage(c *gin.Context, msg string) {
	ok(c, gin.H{"message": i18n.Msg(i18n.FromContext(c), msg)})
}

func fail(c *gin.Context, status int, msg string) {
	// 文案按本次请求协商出的语言返回（Accept-Language → 语言 cookie → 中文，见 internal/i18n）
	c.AbortWithStatusJSON(status, gin.H{"error": i18n.Msg(i18n.FromContext(c), msg)})
}

func badRequest(c *gin.Context, msg string) { fail(c, http.StatusBadRequest, msg) }

// invalidate 写操作后清理相关缓存前缀
func invalidate(prefixes ...string) {
	for _, p := range prefixes {
		cache.DeletePrefix("httpcache:" + p + ":")
	}
}
