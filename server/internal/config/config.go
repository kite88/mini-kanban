package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ErrConfigNotFound 配置文件不存在（首次启动），调用方据此决定是否进入初始化向导。
var ErrConfigNotFound = errors.New("配置文件不存在")

// Duration 支持 "30s" / "1m" 这类时长字符串的 YAML / JSON 解码。
// 不再提供代码内兜底值：格式非法或缺失会在加载配置时直接报错。
type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("时长应为字符串（如 \"30s\"）: %w", err)
	}
	return d.set(s)
}

// set 解析时长文本；留空表示“该配置块未启用”（如 Redis 缓存），非法值仍然报错。
func (d *Duration) set(s string) error {
	if strings.TrimSpace(s) == "" {
		*d = 0
		return nil
	}
	parsed, err := ParseDuration(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) { return d.String(), nil }

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("时长应为字符串（如 \"30s\"）: %w", err)
	}
	return d.set(s)
}

func (d Duration) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }

func (d Duration) String() string { return formatDuration(time.Duration(d)) }

// formatDuration 输出更贴近书写习惯的时长文本：1m0s -> 1m、1h0m0s -> 1h。
func formatDuration(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "0s") && len(s) > 2 {
		if trimmed := strings.TrimSuffix(s, "0s"); isUnit(trimmed[len(trimmed)-1], 'm', 'h') {
			s = trimmed
		}
	}
	if strings.HasSuffix(s, "0m") && len(s) > 2 {
		if trimmed := strings.TrimSuffix(s, "0m"); isUnit(trimmed[len(trimmed)-1], 'h', 'd') {
			s = trimmed
		}
	}
	return s
}

func isUnit(last byte, units ...byte) bool {
	for _, u := range units {
		if last == u {
			return true
		}
	}
	return false
}

// ParseDuration 解析时长字符串，空值 / 非法值一律报错（不做兜底）。
func ParseDuration(s string) (Duration, error) {
	parsed, err := time.ParseDuration(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("时长格式无效 %q: %w", s, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("时长必须大于 0，当前为 %q", s)
	}
	return Duration(parsed), nil
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port string `yaml:"port" json:"port"`
	Mode string `yaml:"mode" json:"mode"` // debug | release | test
}

// DatabaseConfig PostgreSQL 配置；dsn 非空时优先使用，否则按字段拼装
type DatabaseConfig struct {
	DSN                  string `yaml:"dsn" json:"dsn"`
	Host                 string `yaml:"host" json:"host"`
	Port                 int    `yaml:"port" json:"port"`
	User                 string `yaml:"user" json:"user"`
	Password             string `yaml:"password" json:"password"`
	DBName               string `yaml:"dbname" json:"dbname"`
	SSLMode              string `yaml:"sslmode" json:"sslmode"`
	PreferSimpleProtocol bool   `yaml:"preferSimpleProtocol" json:"preferSimpleProtocol"` // 关闭扩展查询协议，兼容部分驱动/数据库版本组合
	LogLevel             string `yaml:"logLevel" json:"logLevel"`                         // gorm 日志级别：silent / error / warn / info，仅由配置文件控制
}

func (d DatabaseConfig) DSNString() string {
	if strings.TrimSpace(d.DSN) != "" {
		return d.DSN
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, url.QueryEscape(d.Password), d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// RedisConfig Redis 及缓存时长配置
type RedisConfig struct {
	Host         string   `yaml:"host" json:"host"`
	Port         int      `yaml:"port" json:"port"`
	Password     string   `yaml:"password" json:"password"`
	DB           int      `yaml:"db" json:"db"`
	CacheTTL     Duration `yaml:"cacheTTL" json:"cacheTTL"`         // 项目 / 任务列表缓存时长
	UserCacheTTL Duration `yaml:"userCacheTTL" json:"userCacheTTL"` // 用户列表缓存时长
}

func (r RedisConfig) Addr() string { return fmt.Sprintf("%s:%d", r.Host, r.Port) }

func (r RedisConfig) TTL() time.Duration { return time.Duration(r.CacheTTL) }

func (r RedisConfig) UserTTL() time.Duration { return time.Duration(r.UserCacheTTL) }

type JWTConfig struct {
	Secret   string `yaml:"secret" json:"secret"`
	TTLHours int    `yaml:"ttlHours" json:"ttlHours"`
}

// SeedConfig 首次启动（用户表为空）时的初始管理员账号配置。
// 初始管理员写入成功后，这几个字段会由 ClearSeed 自动清空（enabled 置为 false）。
type SeedConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`             // 是否写入初始管理员账号；false 则不写入任何账号
	ForceReset    bool   `yaml:"forceReset" json:"forceReset"`       // 是否无视「用户表为空」的前提，强制重置 / 重建该管理员账号
	AdminEmail    string `yaml:"adminEmail" json:"adminEmail"`       // 初始管理员邮箱（登录账号）
	AdminUsername string `yaml:"adminUsername" json:"adminUsername"` // 初始管理员显示名
	AdminPassword string `yaml:"adminPassword" json:"adminPassword"` // 初始管理员密码
}

// StorageConfig 任务附件与富文本内嵌图片的存储配置。
// 与 database.logLevel 一样属于「只在配置文件里维护」的配置段（初始化向导不展示，
// 向导写配置时按模板原值保留）。
type StorageConfig struct {
	Dir       string `yaml:"dir" json:"dir"`             // 存放目录，相对路径按后端工作目录解析
	MaxSizeMB int    `yaml:"maxSizeMB" json:"maxSizeMB"` // 单个文件大小上限（MB）
}

type Config struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Database DatabaseConfig `yaml:"database" json:"database"`
	Redis    RedisConfig    `yaml:"redis" json:"redis"`
	JWT      JWTConfig      `yaml:"jwt" json:"jwt"`
	Seed     SeedConfig     `yaml:"seed" json:"seed"`
	Storage  StorageConfig  `yaml:"storage" json:"storage"`
}

// Path 返回配置文件的读取 / 写入位置：CONFIG_FILE 优先，否则工作目录下的 config.yaml。
func Path() string {
	if p := strings.TrimSpace(os.Getenv("CONFIG_FILE")); p != "" {
		return p
	}
	return "config.yaml"
}

// Load 读取配置文件（必填），并用环境变量覆盖（PORT / DB_DSN / REDIS_ADDR / REDIS_PASSWORD / JWT_SECRET）。
// 配置文件缺失时返回 ErrConfigNotFound；内容非法时返回具体错误。任何情况下都不使用代码内兜底值。
func Load() (*Config, error) {
	path, err := FindConfigFile()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %s 失败：%w", path, err)
	}

	cfg, err := Parse(b)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件 %s 失败：%w", path, err)
	}

	if err := applyEnv(cfg); err != nil {
		return nil, fmt.Errorf("环境变量覆盖配置失败：%w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置文件 %s 无效：%w", path, err)
	}
	return cfg, nil
}

// Parse 仅解析配置内容，不做环境变量覆盖与校验（供初始化向导校验表单使用）。
func Parse(b []byte) (*Config, error) {
	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// FindConfigFile 定位配置文件：CONFIG_FILE 环境变量优先，其次在常见位置搜索；找不到返回 ErrConfigNotFound。
func FindConfigFile() (string, error) {
	if p := strings.TrimSpace(os.Getenv("CONFIG_FILE")); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("CONFIG_FILE 指定的配置文件 %s 不可用：%w", p, err)
		}
		return p, nil
	}

	for _, p := range []string{
		"config.yaml",
		"config.yml",
		"configs/config.yaml",
		"server/config.yaml",
		"../config.yaml",
		"../server/config.yaml",
	} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("%w，请复制 config.example.yaml 为 config.yaml，或用 CONFIG_FILE 指定路径", ErrConfigNotFound)
}

// Sanitize 清空敏感字段，避免已保存的密码 / 密钥被回显到初始化向导页面。
func (c *Config) Sanitize() *Config {
	c.Database.Password = ""
	c.Redis.Password = ""
	c.JWT.Secret = ""
	c.Seed.AdminPassword = ""
	return c
}

// InheritUnexposed 用模板（config.example.yaml）里的值补齐「初始化向导不展示」的配置段。
//
// 向导表单只包含 server / database / redis / jwt / seed；storage 与 database.logLevel
// 刻意不进表单（见 config.example.yaml 的注释，它们只由配置文件维护），
// 因此向导提交上来的对象里这两处是空的，直接 Validate 会误报「未配置」而无法保存。
func (c *Config) InheritUnexposed(tpl *Config) {
	if strings.TrimSpace(c.Storage.Dir) == "" {
		c.Storage.Dir = tpl.Storage.Dir
	}
	if c.Storage.MaxSizeMB <= 0 {
		c.Storage.MaxSizeMB = tpl.Storage.MaxSizeMB
	}
	if strings.TrimSpace(c.Database.LogLevel) == "" {
		c.Database.LogLevel = tpl.Database.LogLevel
	}
}

// Validate 校验配置项完整性：所有必需项都取自配置文件，缺失即报错。
func (c *Config) Validate() error {
	var errs []string
	require := func(missing bool, key string) {
		if missing {
			errs = append(errs, key+" 未配置")
		}
	}

	require(strings.TrimSpace(c.Server.Port) == "", "server.port")
	switch c.Server.Mode {
	case "debug", "release", "test":
	default:
		errs = append(errs, "server.mode 必须是 debug / release / test")
	}

	if strings.TrimSpace(c.Database.DSN) == "" {
		require(strings.TrimSpace(c.Database.Host) == "", "database.host")
		require(c.Database.Port <= 0, "database.port")
		require(strings.TrimSpace(c.Database.User) == "", "database.user")
		require(strings.TrimSpace(c.Database.DBName) == "", "database.dbname")
		require(strings.TrimSpace(c.Database.SSLMode) == "", "database.sslmode")
	}

	// Redis 为可选组件：host 留空表示不启用缓存，此时不再要求其余字段
	if strings.TrimSpace(c.Redis.Host) != "" {
		require(c.Redis.Port <= 0, "redis.port")
		require(c.Redis.CacheTTL <= 0, "redis.cacheTTL")
		require(c.Redis.UserCacheTTL <= 0, "redis.userCacheTTL")
	}

	require(strings.TrimSpace(c.JWT.Secret) == "", "jwt.secret")
	require(c.JWT.TTLHours <= 0, "jwt.ttlHours")

	// 强制重置管理员（seed.forceReset）同样要求三项齐全
	if c.Seed.Enabled || c.Seed.ForceReset {
		require(strings.TrimSpace(c.Seed.AdminEmail) == "", "seed.adminEmail")
		require(strings.TrimSpace(c.Seed.AdminUsername) == "", "seed.adminUsername")
		require(c.Seed.AdminPassword == "", "seed.adminPassword")
	}

	require(strings.TrimSpace(c.Storage.Dir) == "", "storage.dir")
	require(c.Storage.MaxSizeMB <= 0, "storage.maxSizeMB")

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "；"))
	}
	return nil
}

func applyEnv(c *Config) error {
	if v := os.Getenv("PORT"); v != "" {
		c.Server.Port = v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		c.Database.DSN = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		i := strings.LastIndex(v, ":")
		if i <= 0 {
			return fmt.Errorf("REDIS_ADDR 格式应为 host:port，当前为 %q", v)
		}
		port, err := strconv.Atoi(strings.TrimSpace(v[i+1:]))
		if err != nil || port <= 0 {
			return fmt.Errorf("REDIS_ADDR 端口无效，当前为 %q", v)
		}
		c.Redis.Host = v[:i]
		c.Redis.Port = port
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.JWT.Secret = v
	}
	return nil
}
