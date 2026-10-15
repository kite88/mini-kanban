// Package router 集中注册 HTTP 路由与中间件。
package router

import (
	"github.com/gin-gonic/gin"

	"mini-kanban/internal/cache"
	"mini-kanban/internal/config"
	"mini-kanban/internal/handlers"
	"mini-kanban/internal/i18n"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/setup"
)

// New 组装全部路由与中间件。
// 缓存时长来自配置（redis.cacheTTL / redis.userCacheTTL）。
// webApp 表示 web-dist 目录里有可用页面（发布包形态），只影响 /setup 完成页是否给出首页入口。
func New(cfg *config.Config, webApp bool) *gin.Engine {
	r := gin.New()
	// 语言协商放在最前面：后面的处理器与鉴权中间件都按它返回中文 / 英文文案
	r.Use(gin.Recovery(), i18n.Middleware(), middleware.Logger(), middleware.CORS())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "redis": cache.Enabled()})
	})

	// 已配置完成时访问 /setup 给出“服务运行正常”的反馈页
	r.GET("/setup", setup.DoneHandler(cfg, webApp))

	api := r.Group("/api")
	{
		api.POST("/auth/login", handlers.Login)
		// 附件下载：富文本里的 <img src> 与附件链接由浏览器直接请求，带不了 Authorization 头，
		// 只能把凭据放进 URL；因此这里只认下载票据（aud=file，见 POST /files/ticket），
		// 登录 token 不接受，避免下载地址泄露等价于登录态泄露。
		api.GET("/files/:name", middleware.FileTicketRequired(), handlers.ServeFile)

		auth := api.Group("", middleware.AuthRequired())
		auth.GET("/auth/me", handlers.Me)
		auth.POST("/auth/logout", handlers.Logout)
		auth.PUT("/auth/profile", handlers.UpdateProfile)

		// 部门管理：查看对所有登录用户开放，写操作仅管理员
		departments := auth.Group("/departments")
		departments.GET("", middleware.Cache(cfg.Redis.UserTTL(), "departments"), handlers.ListDepartments)
		departments.GET("/:id", middleware.Cache(cfg.Redis.UserTTL(), "departments"), handlers.GetDepartment)
		departments.POST("", middleware.AdminRequired(), handlers.CreateDepartment)
		departments.PUT("/:id", middleware.AdminRequired(), handlers.UpdateDepartment)
		departments.DELETE("/:id", middleware.AdminRequired(), handlers.DeleteDepartment)

		// 用户管理：查看对所有登录用户开放，写操作仅管理员
		users := auth.Group("/users")
		users.GET("", middleware.Cache(cfg.Redis.UserTTL(), "users"), handlers.ListUsers)
		users.GET("/:id", middleware.Cache(cfg.Redis.UserTTL(), "users"), handlers.GetUser)
		users.POST("", middleware.AdminRequired(), handlers.CreateUser)
		users.PUT("/:id", middleware.AdminRequired(), handlers.UpdateUser)
		users.DELETE("/:id", middleware.AdminRequired(), handlers.DeleteUser)

		projects := auth.Group("/projects")
		projects.GET("", middleware.Cache(cfg.Redis.TTL(), "projects"), handlers.ListProjects)
		projects.POST("", handlers.CreateProject)
		projects.GET("/:id", middleware.Cache(cfg.Redis.TTL(), "projects"), handlers.GetProject)
		projects.PUT("/:id", handlers.UpdateProject)
		projects.DELETE("/:id", handlers.DeleteProject)
		projects.GET("/:id/tasks", middleware.Cache(cfg.Redis.TTL(), "tasks"), handlers.ListTasks)
		projects.POST("/:id/tasks", handlers.CreateTask)

		// 项目资料（需求文档 / 会议纪要 / 设计稿 / 项目密钥 / 测试账密）：挂在项目下，表单为标题 + 富文本 + 附件
		projects.GET("/:id/docs", middleware.Cache(cfg.Redis.TTL(), "docs"), handlers.ListProjectDocs)
		projects.POST("/:id/docs", handlers.CreateProjectDoc)

		docs := auth.Group("/docs")
		docs.PUT("/:id", handlers.UpdateProjectDoc)
		docs.DELETE("/:id", handlers.DeleteProjectDoc)

		// 附件上传（登录态）：multipart/form-data，一次可传多个文件
		auth.POST("/uploads", handlers.UploadFiles)

		// 附件下载票据：登录后换取一张短时效、只能读附件的凭据，供下载地址拼 ?ticket=
		auth.POST("/files/ticket", handlers.IssueFileTicket)

		tasks := auth.Group("/tasks")
		tasks.GET("/:id", middleware.Cache(cfg.Redis.TTL(), "tasks"), handlers.GetTask)
		tasks.PUT("/:id", handlers.UpdateTask)
		tasks.PATCH("/:id", handlers.UpdateTask)
		tasks.DELETE("/:id", handlers.DeleteTask)
		tasks.PATCH("/:id/move", handlers.MoveTask)

		// 回收箱：软删除数据的集中管理（列表 / 恢复 / 彻底删除）。
		// 列表刻意不挂缓存中间件：回收箱是低频入口，实时性比缓存收益更重要。
		trash := auth.Group("/trash")
		trash.GET("", handlers.ListTrash)
		trash.POST("/projects/:id/restore", handlers.RestoreProject)
		trash.DELETE("/projects/:id", handlers.PurgeProject)
		trash.POST("/tasks/:id/restore", handlers.RestoreTask)
		trash.DELETE("/tasks/:id", handlers.PurgeTask)
		trash.POST("/docs/:id/restore", handlers.RestoreDoc)
		trash.DELETE("/docs/:id", handlers.PurgeDoc)

		// 操作日志：涵盖项目 / 任务 / 项目资料 / 用户 / 部门，物理删除的条目带数据快照。
		// 同样不挂缓存：日志是审计用途，且含用户与部门的敏感动作，仅管理员可见。
		logs := auth.Group("/logs", middleware.AdminRequired())
		logs.GET("", handlers.ListLogs)
		logs.GET("/:id", handlers.GetLog)
	}

	return r
}
