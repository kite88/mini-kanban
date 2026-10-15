package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"mini-kanban/internal/cache"
	"mini-kanban/internal/config"
	"mini-kanban/internal/db"
	"mini-kanban/internal/models"
	"mini-kanban/internal/router"
	"mini-kanban/internal/setup"
	"mini-kanban/internal/storage"
	"mini-kanban/internal/utils"
)

// configTemplate 配置模板，用作初始化向导写入 config.yaml 时的底稿（保留注释与排版）
//
//go:embed config.example.yaml
var configTemplate []byte

func main() {
	cfg, err := loadOrSetup()
	if err != nil {
		log.Fatalf("%v", err)
	}
	serve(cfg)
}

// loadOrSetup 加载配置。
// 未检测到配置文件（首次启动），或显式开启 SETUP_FORCE 时，先进入初始化配置向导，
// 待 config.yaml 写入完成后返回配置，随后按正常流程启动业务服务。
func loadOrSetup() (*config.Config, error) {
	cfg, err := config.Load()
	force := isTruthy(os.Getenv("SETUP_FORCE"))

	if err == nil && !force {
		return cfg, nil
	}
	if err != nil && !errors.Is(err, config.ErrConfigNotFound) {
		if !force {
			return nil, err
		}
		log.Printf("现有配置存在问题（SETUP_FORCE 已开启，将在向导中重新配置）：%v", err)
	}

	prefill := &config.Config{}
	if cfg != nil {
		// 重新配置：用现有配置预填表单（密码与密钥不会回显）
		prefill = cfg
	} else if parsed, perr := config.Parse(configTemplate); perr == nil {
		// 首次启动：用模板中的示例值预填表单（仅用于表单，不参与运行时兜底）
		prefill = parsed
	}

	if err := setup.Run(setup.Options{
		Template:   configTemplate,
		Prefill:    prefill.Sanitize(),
		Port:       setupPort(),
		ConfigPath: config.Path(),
		Force:      force,
	}); err != nil {
		return nil, err
	}

	cfg, err = config.Load()
	if err != nil {
		return nil, fmt.Errorf("初始化配置已写入，但加载失败：%w", err)
	}
	return cfg, nil
}

// setupPort 初始化向导自身的监听端口：SETUP_PORT 环境变量优先，默认 8281
// （此时还没有配置文件，无法从配置中读取端口）。
func setupPort() string {
	if v := strings.TrimSpace(os.Getenv("SETUP_PORT")); v != "" {
		return v
	}
	return "8281"
}

func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func serve(cfg *config.Config) {
	gin.SetMode(cfg.Server.Mode)

	// 雪花节点号：单机固定 0；多实例部署时用 SNOWFLAKE_NODE_ID（0~1023）区分机器。
	// 必须在 db.Init 之前，因为启动过程里可能就有写库（seed 建管理员）。
	nodeID := int64(0)
	if v := strings.TrimSpace(os.Getenv("SNOWFLAKE_NODE_ID")); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 0 || n > 1023 {
			log.Fatalf("SNOWFLAKE_NODE_ID 必须是 0~1023 之间的整数，当前为 %q", v)
		}
		nodeID = n
	}
	if err := models.InitID(nodeID); err != nil {
		log.Fatal(err)
	}

	db.Init(cfg)
	// Redis 为可选组件：redis.host 留空表示不启用缓存
	if strings.TrimSpace(cfg.Redis.Host) == "" {
		log.Printf("未配置 Redis（redis.host 留空），缓存功能已关闭")
	} else {
		cache.Init(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	}
	utils.Init(cfg)
	// 附件存储目录（storage.dir / storage.maxSizeMB）
	if err := storage.Init(cfg); err != nil {
		log.Fatalf("%v", err)
	}

	// 前端静态站点：web-dist 目录（发布时随可执行文件同目录分发，也可用 WEB_DIST_DIR 指定），
	// 未匹配到 /api、/healthz、/setup 的请求都会落到它（含 SPA 路由兜底）。
	// 目录缺失时自动降级为「只提供接口」，HasWebApp 同样决定 /setup 完成页是否给出首页入口。
	ResolveWebDist()
	r := router.New(cfg, HasWebApp())
	r.NoRoute(serveWebApp)

	log.Printf("mini-kanban 启动于 :%s (Postgres: %s/%s, Redis: %s, 缓存: %v, 前端: %v)",
		cfg.Server.Port, cfg.Database.Host, cfg.Database.DBName, cfg.Redis.Addr(), cache.Enabled(), webAppDesc())
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
