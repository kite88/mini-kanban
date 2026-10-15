package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mini-kanban/internal/db"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/models"
	"mini-kanban/internal/oplog"
	"mini-kanban/internal/utils"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "邮箱和密码不能为空")
		return
	}
	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		fail(c, http.StatusUnauthorized, "邮箱或密码错误")
		return
	}
	if !user.CheckPassword(req.Password) {
		fail(c, http.StatusUnauthorized, "邮箱或密码错误")
		return
	}
	token, _, err := utils.GenerateToken(&user)
	if err != nil {
		fail(c, http.StatusInternalServerError, "令牌生成失败")
		return
	}
	ok(c, gin.H{"token": token, "user": user})
}

func Me(c *gin.Context) {
	ok(c, middleware.CurrentUser(c))
}

func Logout(c *gin.Context) {
	if claims, exists := c.Get(middleware.CtxClaims); exists {
		cl := claims.(*utils.Claims)
		exp := time.Now()
		if cl.RegisteredClaims.ExpiresAt != nil {
			exp = cl.RegisteredClaims.ExpiresAt.Time
		}
		middleware.BlacklistToken(cl.JTI, exp)
	}
	okMessage(c, "已登出")
}

// UpdateProfile 修改当前登录用户的资料（不含角色）
func UpdateProfile(c *gin.Context) {
	u := middleware.CurrentUser(c)
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	// 保留修改前的内容，供操作日志对比（值拷贝，互不影响）
	before := *u
	if req.Username != "" {
		u.Username = req.Username
	}
	if email := utils.NormalizeEmail(req.Email); email != "" {
		if !utils.IsEmail(email) {
			badRequest(c, "邮箱格式不正确")
			return
		}
		u.Email = email
	}
	if req.Password != "" {
		_ = u.SetPassword(req.Password)
	}
	if err := db.DB.Save(u).Error; err != nil {
		fail(c, http.StatusBadRequest, "保存失败: "+err.Error())
		return
	}
	invalidate("users", "projects", "tasks")
	// 改自己的资料也进用户日志（只记改了什么，不记密码内容）
	summary := oplog.ChangesSummary(
		oplog.Change{Label: "用户名", From: before.Username, To: u.Username},
		oplog.Change{Label: "邮箱", From: before.Email, To: u.Email},
	)
	if req.Password != "" {
		summary = oplog.AppendSummary(summary, "密码已重置")
	}
	oplog.Record(u, oplog.Entry{
		Action:     models.ActionUpdate,
		TargetType: models.TargetUser,
		TargetID:   u.ID.String(),
		TargetName: u.Username,
		Summary:    oplog.AppendSummary(summary, "本人修改个人资料"),
	})
	ok(c, u)
}
