package db

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"mini-kanban/internal/config"
	"mini-kanban/internal/models"
)

var DB *gorm.DB

// tabler 用于从模型取表名：裸 SQL 里的表名统一走模型的 TableName，
// 表名调整时不必再四处改字符串。
type tabler interface{ TableName() string }

func tableName(m tabler) string { return m.TableName() }

// gormLogLevels 配置项 database.logLevel 到 gorm 日志级别的映射。
var gormLogLevels = map[string]logger.LogLevel{
	"silent": logger.Silent,
	"error":  logger.Error,
	"warn":   logger.Warn,
	"info":   logger.Info,
}

func Init(cfg *config.Config) {
	level, ok := gormLogLevels[cfg.Database.LogLevel]
	if !ok {
		log.Fatalf("database.logLevel 无效: %q（可选 silent / error / warn / info）", cfg.Database.LogLevel)
	}

	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.Database.DSNString(),
		PreferSimpleProtocol: cfg.Database.PreferSimpleProtocol,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	migrate()
	seed(cfg)
}

// migrate 建表并补齐缺失列。
// 注：gorm 的 AutoMigrate 会调用 ColumnTypes，在 PostgreSQL 18 + pgx(<=5.6) 组合下
// 会返回 "insufficient arguments"，这里改用 HasTable/CreateTable + HasColumn/AddColumn 规避。
func migrate() {
	ensureIDSchema()

	// 首次上线 project_members 时，把已有项目初始化为「全员可见」，
	// 避免新规则生效后历史项目突然对所有人不可见（仅执行一次）。
	projectMembersIsNew := !DB.Migrator().HasTable(&models.ProjectMember{})

	// 顺序即建表顺序：外键指向的表要先建（users 关联 departments；资料附件关联资料）
	// operation_logs 无外键（操作人只用 id + 名称快照，不挂关联），放在最后
	for _, m := range []interface{}{&models.Department{}, &models.User{}, &models.Project{}, &models.Task{}, &models.ProjectMember{}, &models.TaskAttachment{}, &models.ProjectDoc{}, &models.ProjectDocAttachment{}, &models.UploadedFile{}, &models.OperationLog{}} {
		mig := DB.Migrator()
		if !mig.HasTable(m) {
			if err := mig.CreateTable(m); err != nil {
				log.Fatalf("创建表失败: %v", err)
			}
			continue
		}

		s, err := schema.Parse(m, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			log.Fatalf("解析模型失败: %v", err)
		}
		for _, f := range s.Fields {
			if f.DBName == "" || f.IgnoreMigration {
				continue
			}
			if !mig.HasColumn(m, f.DBName) {
				if err := mig.AddColumn(m, f.DBName); err != nil {
					log.Printf("添加列 %s 失败: %v", f.DBName, err)
				}
			}
		}
	}

	// 软删除列对老表是新增列：上面的 AddColumn 只补列不补索引，
	// 这里按需建索引（查询都会带 deleted_at IS NULL，没索引会退化成全表扫描）。
	for _, m := range []interface{}{&models.Project{}, &models.Task{}, &models.TaskAttachment{}, &models.ProjectDoc{}, &models.ProjectDocAttachment{}} {
		t, isTabler := m.(tabler)
		if !isTabler {
			continue
		}
		name := "idx_" + tableName(t) + "_deleted_at"
		if !DB.Migrator().HasIndex(m, name) {
			if err := DB.Migrator().CreateIndex(m, name); err != nil {
				log.Printf("创建索引 %s 失败: %v", name, err)
			}
		}
	}

	// 注：旧版本里有一段「把任务的空 creator_id 回退为项目创建者」的历史数据补齐 SQL，
	// ID 改造后 creator_id 是雪花整数、project_id 是 uuid，该 SQL 已无意义（且 COALESCE(x,0)=0
	// 这种整数哨兵判断对 uuid 不成立），故随本次改造一并移除。

	projects := tableName(&models.Project{})
	if projectMembersIsNew {
		projectMembers := tableName(&models.ProjectMember{})
		if err := DB.Exec(fmt.Sprintf(`INSERT INTO %s (project_id, user_id, created_at)
			SELECT p.id, u.id, NOW() FROM %s p CROSS JOIN %s u`,
			projectMembers, projects, tableName(&models.User{}))).Error; err != nil {
			log.Printf("初始化项目成员失败: %v", err)
		} else {
			log.Printf("已初始化 %s：现有项目默认对所有用户可见", projectMembers)
		}
	}
}

// ensureIDSchema 拦住旧结构的库。
//
// 主键约定是「users 用雪花 bigint、其他表用 uuid」：users.id 新旧都是 bigint，但
// 其他表的主键要从 bigserial 变成 uuid —— 这是补列迁移做不到的列类型变更，
// 所以检测到老结构时直接给出删库指引，而不是让它带着旧列类型跑起来。
func ensureIDSchema() {
	if !DB.Migrator().HasTable(&models.Project{}) {
		return // 全新库，按新结构建表即可
	}
	table := tableName(&models.Project{})
	var dataType string
	row := DB.Raw(`SELECT data_type FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = ? AND column_name = 'id'`, table).Row()
	if err := row.Scan(&dataType); err != nil {
		return // 查询失败不阻断启动，交给后续建表逻辑报错
	}
	if dataType == "uuid" {
		return
	}
	log.Fatalf(`数据库仍是旧结构（%s.id 类型为 %s，应为 uuid）。
ID 已改造为「用户表雪花 ID + 其他表 UUID v7」，无法通过补列迁移，请删库重置后重启：
  DROP SCHEMA public CASCADE;
  CREATE SCHEMA public;`, table, dataType)
}

// seed 按 seed 配置准备初始管理员账号，两种模式：
//   - seed.enabled：仅在「用户表为空」时创建，已有用户则跳过（首次部署用）；
//   - seed.forceReset：无视上述前提，把 AdminEmail 对应的账号重置为管理员并更新密码，
//     该邮箱不存在则新建（忘记管理员密码、又不便清库重来时用）。
//
// 两种模式成功后都会清空配置文件里的 seed 段（开关置否、账号与密码置空），
// 避免明文密码长期留在磁盘上。
func seed(cfg *config.Config) {
	if cfg.Seed.ForceReset {
		forceResetAdmin(cfg)
		return
	}
	if !cfg.Seed.Enabled {
		return
	}

	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	admin := models.User{
		Email:    strings.TrimSpace(cfg.Seed.AdminEmail),
		Username: strings.TrimSpace(cfg.Seed.AdminUsername),
		Role:     models.RoleAdmin,
	}
	if err := admin.SetPassword(cfg.Seed.AdminPassword); err != nil {
		log.Printf("初始管理员密码加密失败: %v", err)
		return
	}
	if err := DB.Create(&admin).Error; err != nil {
		log.Printf("创建初始管理员失败: %v", err)
		return
	}
	log.Printf("已创建初始管理员：%s", admin.Email)

	clearSeedAndLog()
}

// forceResetAdmin 确保 cfg.Seed.AdminEmail 对应的账号存在、角色为管理员、
// 且密码为配置里的值；该邮箱不存在则新建。
// 与 seed 的「仅用户表为空时创建」不同，这里不看库里有没有数据。
func forceResetAdmin(cfg *config.Config) {
	email := strings.TrimSpace(cfg.Seed.AdminEmail)
	username := strings.TrimSpace(cfg.Seed.AdminUsername)

	var user models.User
	err := DB.Where("email = ?", email).First(&user).Error

	switch {
	case err == nil:
		if perr := user.SetPassword(cfg.Seed.AdminPassword); perr != nil {
			log.Printf("重置管理员密码失败（加密出错）：%v", perr)
			return
		}
		user.Username = username
		user.Role = models.RoleAdmin
		if uerr := DB.Save(&user).Error; uerr != nil {
			log.Printf("重置管理员密码失败（写库出错）：%v", uerr)
			return
		}
		log.Printf("已重置管理员账号：%s（用户名与 admin 角色一并确保）", email)
	case errors.Is(err, gorm.ErrRecordNotFound):
		admin := models.User{Email: email, Username: username, Role: models.RoleAdmin}
		if perr := admin.SetPassword(cfg.Seed.AdminPassword); perr != nil {
			log.Printf("创建管理员失败（加密出错）：%v", perr)
			return
		}
		if cerr := DB.Create(&admin).Error; cerr != nil {
			log.Printf("创建管理员失败（写库出错）：%v", cerr)
			return
		}
		log.Printf("该邮箱原本不存在，已新建管理员账号：%s", email)
	default:
		log.Printf("查询管理员账号失败：%v", err)
		return
	}

	clearSeedAndLog()
}

// clearSeedAndLog 清空配置文件中的 seed 段并记录结果，避免明文密码留在磁盘上。
func clearSeedAndLog() {
	path, err := config.ClearSeed()
	if err != nil {
		log.Printf("清空配置文件 seed 段失败：%v（请手动把 seed.enabled / seed.forceReset 改回 false，并清空 adminEmail / adminUsername / adminPassword）", err)
		return
	}
	log.Printf("已清空 %s 中的 seed 段：开关置 false，账号与密码已置空", path)
}
