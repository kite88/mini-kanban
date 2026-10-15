package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mini-kanban/internal/cache"
	"mini-kanban/internal/db"
	"mini-kanban/internal/i18n"
	"mini-kanban/internal/models"
	"mini-kanban/internal/utils"
)

const (
	CtxUserID = "userID"
	CtxUser   = "user"
	CtxClaims = "claims"
)

// currentUserID 当前登录用户的 id（雪花 ID 的数值形态，0 表示未登录）
func currentUserID(c *gin.Context) int64 {
	if v, ok := c.Get(CtxUserID); ok {
		return v.(int64)
	}
	return 0
}

// abortUnauthorized 鉴权失败：显式 no-store，避免浏览器把 401 缓存住导致换票据后仍不重发
func abortUnauthorized(c *gin.Context, msg string) {
	c.Header("Cache-Control", "no-store")
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.FromContext(c), msg)})
}

// resolveSession 校验会话是否仍然有效（登出黑名单 + 用户存在），通过则装载登录上下文
func resolveSession(c *gin.Context, claims *utils.Claims) bool {
	if cache.Enabled() {
		if _, ok := cache.Get(blacklistKey(claims.JTI)); ok {
			abortUnauthorized(c, "登录已失效")
			return false
		}
	}

	var user models.User
	if err := db.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		abortUnauthorized(c, "用户不存在")
		return false
	}
	c.Set(CtxUserID, int64(user.ID))
	c.Set(CtxUser, &user)
	c.Set(CtxClaims, claims)
	return true
}

// AuthRequired 校验请求头里的 Bearer Token（登录 token），并支持登出黑名单（Redis）。
// 下载票据（aud=file）用的是同一个签名密钥，必须显式拒绝，
// 否则拿到票据就等于拿到登录态、能调用全部接口。
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			abortUnauthorized(c, "未登录")
			return
		}
		claims, err := utils.ParseToken(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			abortUnauthorized(c, "登录已失效")
			return
		}
		if claims.IsFileTicket() {
			abortUnauthorized(c, "登录已失效")
			return
		}
		if !resolveSession(c, claims) {
			return
		}
		c.Next()
	}
}

// FileTicketRequired 附件下载鉴权：只认 ?ticket=<下载票据>。
// 下载地址由浏览器直接发起（<img src> / <a href>），带不了 Authorization 头，凭据只能进 URL；
// 这里刻意不接受登录 token，因此 URL 即使外泄，泄露的也只是一张短时效、
// 只能读附件的票据，而不是可以调用全部接口的登录凭据。
// 票据走独立派生密钥验签（utils.ParseFileTicket），aud 亦强制为 file。
func FileTicketRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(c.Query("ticket"))
		if raw == "" {
			abortUnauthorized(c, "缺少下载凭证")
			return
		}
		claims, err := utils.ParseFileTicket(raw)
		if err != nil || !claims.IsFileTicket() {
			abortUnauthorized(c, "下载凭证无效")
			return
		}
		if !resolveSession(c, claims) {
			return
		}
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := c.Get(CtxUser)
		if !ok || u.(*models.User).Role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": i18n.Msg(i18n.FromContext(c), "需要管理员权限"),
			})
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) *models.User {
	u, _ := c.Get(CtxUser)
	if u == nil {
		return nil
	}
	return u.(*models.User)
}

func blacklistKey(jti string) string { return "auth:blacklist:" + jti }

// BlacklistToken 将 token 加入黑名单直到其自然过期
func BlacklistToken(jti string, exp time.Time) {
	ttl := time.Until(exp)
	if ttl <= 0 {
		return
	}
	cache.Set(blacklistKey(jti), []byte("1"), ttl)
}
