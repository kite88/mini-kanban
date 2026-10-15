// Package oplog 记录「项目 / 任务 / 项目资料 / 用户 / 部门」的操作日志。
//
// 两类日志的区别在于是否留数据快照：
//   - 普通操作（新建 / 修改 / 软删除 / 恢复）只记一行动作记录，数据本身仍在库里可查；
//   - 物理删除（删除用户、删除部门、回收箱彻底删除）会额外把被删数据整份序列化进 snapshot，
//     因为这类数据删除后无从追溯（见 snapshot.go）。
//
// 日志写入失败不影响业务结果，只记服务端日志；唯一例外是物理删除，
// 它的删除与快照写在同一个事务里（RecordTx），保证「数据没了但快照也没留下」不会发生。
package oplog

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"mini-kanban/internal/db"
	"mini-kanban/internal/models"
)

// Entry 一条待写入的操作日志
type Entry struct {
	Action     string // models.Action*
	TargetType string // models.Target*
	// TargetID 目标 id：用户 / 部门是雪花数字或 UUID，统一按字符串写入（见 models.OperationLog）
	TargetID   string
	TargetName string // 对象名称：项目标题 / 任务标题 / 用户名 / 部门名
	Summary    string // 变更摘要，可为空
	Snapshot   any    // 数据快照；nil 表示不留快照（只有物理删除需要）
}

// Record 写入一条操作日志
func Record(op *models.User, e Entry) {
	if err := db.DB.Create(newRow(op, e)).Error; err != nil {
		log.Printf("写入操作日志失败（%s %s#%s）：%v", e.TargetType, e.Action, e.TargetID, err)
	}
}

// RecordTx 在调用方事务内写入操作日志：物理删除与它的快照同生共死
func RecordTx(tx *gorm.DB, op *models.User, e Entry) error {
	return tx.Create(newRow(op, e)).Error
}

func newRow(op *models.User, e Entry) *models.OperationLog {
	row := &models.OperationLog{
		Action:     e.Action,
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		TargetName: Truncate(e.TargetName, 200),
		Summary:    Truncate(e.Summary, 500),
	}
	// 操作人只留 id + 用户名快照：用户被删除后日志依旧能说明「是谁做的」
	if op != nil {
		row.OperatorID = op.ID
		row.OperatorName = Truncate(op.Username, 64)
	}
	if e.Snapshot != nil {
		data, err := json.Marshal(e.Snapshot)
		if err != nil {
			log.Printf("序列化数据快照失败（%s#%s）：%v", e.TargetType, e.TargetID, err)
		} else {
			row.Snapshot = data
		}
	}
	return row
}

// Change 一个字段的变更前后值，用于拼装 update 类日志的摘要。
// From / To 传可直接展示的值（字符串、时间、nil），不要传富文本正文。
type Change struct {
	Label string
	From  any
	To    any
}

// ChangesSummary 生成「标题：旧 → 新；状态：待办 → 进行中」形式的摘要，未变化的字段自动跳过
func ChangesSummary(changes ...Change) string {
	parts := make([]string, 0, len(changes))
	for _, c := range changes {
		from, to := describe(c.From), describe(c.To)
		if from == to {
			continue
		}
		parts = append(parts, c.Label+"："+from+" → "+to)
	}
	return strings.Join(parts, "；")
}

// AppendSummary 在已有摘要后追加一句（摘要之间用「；」连接）
func AppendSummary(summary, extra string) string {
	switch {
	case summary == "":
		return extra
	case extra == "":
		return summary
	default:
		return summary + "；" + extra
	}
}

const (
	emptyText  = "（空）"
	timeLayout = "2006-01-02 15:04"
)

// describe 把变更值渲染成简短文本：空值统一成「（空）」，过长内容截断成一行
func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return emptyText
	case string:
		if strings.TrimSpace(x) == "" {
			return emptyText
		}
		// 富文本 / 多行内容折成一行，避免摘要里出现换行
		return Truncate(strings.Join(strings.Fields(x), " "), 60)
	case *string:
		if x == nil {
			return emptyText
		}
		return describe(*x)
	case time.Time:
		return x.Format(timeLayout)
	case *time.Time:
		if x == nil {
			return emptyText
		}
		return x.Format(timeLayout)
	case bool:
		if x {
			return "是"
		}
		return "否"
	default:
		return Truncate(fmt.Sprint(v), 60)
	}
}

// Truncate 按字符截断，避免超出列长度导致日志写入失败
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// ============================ 日志展示用的轻量查询 ============================

// UserName 取用户名（id 为空或用户已删除时返回空串）
func UserName(id *models.SnowflakeID) string {
	if id == nil || *id == 0 {
		return ""
	}
	var name string
	db.DB.Model(&models.User{}).Where("id = ?", *id).Pluck("username", &name)
	return name
}

// DepartmentName 取部门名（id 为空时返回空串）
func DepartmentName(id *models.UUID) string {
	if id == nil || *id == (models.UUID{}) {
		return ""
	}
	var name string
	db.DB.Model(&models.Department{}).Where("id = ?", *id).Pluck("name", &name)
	return name
}

// ProjectTitle 取项目标题（含已软删除的项目，快照场景需要）
func ProjectTitle(id models.UUID) string {
	var title string
	db.DB.Unscoped().Model(&models.Project{}).Where("id = ?", id).Pluck("title", &title)
	return title
}

// UserNames 把用户 id 列表渲染成「张三、李四」形式的文本（最多 3 个，超出用「等 N 人」）
func UserNames(ids []models.SnowflakeID) string {
	if len(ids) == 0 {
		return ""
	}
	var names []string
	db.DB.Model(&models.User{}).Where("id IN ?", ids).Pluck("username", &names)
	if len(names) == 0 {
		return ""
	}
	const maxNames = 3
	if len(names) > maxNames {
		return strings.Join(names[:maxNames], "、") + "等 " + strconv.Itoa(len(names)) + " 人"
	}
	return strings.Join(names, "、")
}
