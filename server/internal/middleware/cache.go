package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mini-kanban/internal/cache"
	"mini-kanban/internal/i18n"
)

// Cache 是通用的 GET 响应缓存中间件（Redis）。
// 命中时直接返回缓存的 JSON；写接口通过 cache.DeletePrefix 主动失效。
func Cache(ttl time.Duration, namespace string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet || !cache.Enabled() {
			c.Next()
			return
		}
		// key 里带上语言：缓存的是响应体，万一将来 payload 里出现按语言生成的文案，
		// 也不会出现「中文用户的缓存被英文用户命中」。当前缓存的都是纯数据，属防御性设计。
		key := fmt.Sprintf(
			"httpcache:%s:%s:%d:%s?%s",
			namespace, i18n.FromContext(c), currentUserID(c), c.Request.URL.Path, c.Request.URL.RawQuery,
		)

		if data, ok := cache.Get(key); ok {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.Header("X-Cache", "HIT")
			c.String(http.StatusOK, string(data))
			c.Abort()
			return
		}

		w := &cacheWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = w
		c.Next()

		if w.Status() == http.StatusOK {
			cache.Set(key, w.body.Bytes(), ttl)
		}
	}
}

type cacheWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *cacheWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *cacheWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
