package setup

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mini-kanban/internal/config"
)

const probeTimeout = 5 * time.Second

// ProbeResult 连接测试结果
type ProbeResult struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	Detail    string `json:"detail,omitempty"`
	LatencyMs int64  `json:"latencyMs"`
}

// probeDatabase 用提交的表单参数真实连一次 PostgreSQL（与 db.Init 使用同一套驱动配置）。
func probeDatabase(d config.DatabaseConfig) ProbeResult {
	if strings.TrimSpace(d.DSN) == "" {
		if strings.TrimSpace(d.Host) == "" || d.Port <= 0 || strings.TrimSpace(d.DBName) == "" {
			return ProbeResult{Message: "请先填写 host / port / dbname，或直接填写完整 dsn"}
		}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  d.DSNString(),
		PreferSimpleProtocol: d.PreferSimpleProtocol,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return ProbeResult{Message: "连接失败", Detail: fixEncoding(err.Error()), LatencyMs: elapsed(start)}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return ProbeResult{Message: "连接失败", Detail: fixEncoding(err.Error()), LatencyMs: elapsed(start)}
	}
	defer sqlDB.Close()

	if err := sqlDB.PingContext(ctx); err != nil {
		return ProbeResult{Message: "连接失败，请检查地址 / 账号 / 密码 / 库名", Detail: fixEncoding(err.Error()), LatencyMs: elapsed(start)}
	}

	detail := ""
	var version string
	if err := sqlDB.QueryRowContext(ctx, "select version()").Scan(&version); err == nil {
		detail = firstSentence(version)
	}
	return ProbeResult{OK: true, Message: "连接成功", Detail: detail, LatencyMs: elapsed(start)}
}

// probeRedis 探测 Redis；Redis 仅用于缓存，失败不阻断配置保存。
func probeRedis(r config.RedisConfig) ProbeResult {
	if strings.TrimSpace(r.Host) == "" || r.Port <= 0 {
		return ProbeResult{Message: "请先填写 Redis host / port"}
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:     r.Addr(),
		Password: r.Password,
		DB:       r.DB,
	})
	defer client.Close()

	info, err := client.Ping(ctx).Result()
	if err != nil {
		return ProbeResult{Message: "连接失败，请检查地址 / 密码", Detail: err.Error(), LatencyMs: elapsed(start)}
	}
	return ProbeResult{OK: true, Message: "连接成功（" + info + "）", LatencyMs: elapsed(start)}
}

func elapsed(start time.Time) int64 { return time.Since(start).Milliseconds() }

// fixEncoding 修掉 PostgreSQL 在「连接阶段」下发的非 UTF-8 报错文本。
//
// client_encoding 要等连接建立（选库完成）之后才生效，而认证失败、库不存在
// 这类连接期错误是服务端直接按自己的消息目录编码下发的：中文 Windows 上的
// PostgreSQL（lc_messages = Chinese (Simplified)_China.936）给的是 GBK 字节，
// 在 Go 里就是非法 UTF-8，塞进 JSON 响应会变成一串 �。
//
// 已经是合法 UTF-8 的文本（绝大多数情况，含认证成功后的所有查询报错）原样返回；
// 否则按 GBK 解回来；实在解不干净就丢弃非法字节——保证页面上不出现乱码。
func fixEncoding(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	if out, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), s); err == nil &&
		!strings.ContainsRune(out, utf8.RuneError) {
		return out
	}
	return strings.ToValidUTF8(s, "")
}

func firstSentence(s string) string {
	s = strings.SplitN(strings.TrimSpace(s), ",", 2)[0]
	if len([]rune(s)) > 80 {
		return fmt.Sprintf("%.80s…", s)
	}
	return s
}
