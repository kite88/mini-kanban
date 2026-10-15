package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mini-kanban/internal/db"
	"mini-kanban/internal/models"
)

const (
	defaultLogPageSize = 20
	maxLogPageSize     = 100
)

// 允许的筛选值：不在白名单里的参数按「不限」处理，避免前端拼错参数直接报错
var (
	logTargetTypes = map[string]bool{
		models.TargetProject:    true,
		models.TargetTask:       true,
		models.TargetDoc:        true,
		models.TargetUser:       true,
		models.TargetDepartment: true,
	}
	logActions = map[string]bool{
		models.ActionCreate:  true,
		models.ActionUpdate:  true,
		models.ActionMove:    true,
		models.ActionDelete:  true,
		models.ActionRestore: true,
		models.ActionPurge:   true,
	}
)

// logQuery 操作日志的筛选条件
type logQuery struct {
	targetType string
	action     string
	operatorID models.SnowflakeID
	keyword    string
	page       int
	pageSize   int
}

func queryInt(v string, fallback int) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func parseLogQuery(c *gin.Context) logQuery {
	q := logQuery{
		targetType: c.Query("targetType"),
		action:     c.Query("action"),
		keyword:    strings.TrimSpace(c.Query("keyword")),
		page:       queryInt(c.Query("page"), 1),
		pageSize:   queryInt(c.Query("pageSize"), defaultLogPageSize),
	}
	if !logTargetTypes[q.targetType] {
		q.targetType = ""
	}
	if !logActions[q.action] {
		q.action = ""
	}
	if id, err := models.ParseSnowflakeID(c.Query("operatorId")); err == nil {
		q.operatorID = id
	}
	if q.page < 1 {
		q.page = 1
	}
	if q.pageSize < 1 {
		q.pageSize = defaultLogPageSize
	}
	if q.pageSize > maxLogPageSize {
		q.pageSize = maxLogPageSize
	}
	return q
}

// logScope 按筛选条件构造查询。抽成函数是为了让「统计总数」与「取当页数据」
// 各用一份独立的 statement，避免复用同一个 *gorm.DB 时残留上一次的条件。
func logScope(q logQuery) *gorm.DB {
	tx := db.DB.Model(&models.OperationLog{})
	if q.targetType != "" {
		tx = tx.Where("target_type = ?", q.targetType)
	}
	if q.action != "" {
		tx = tx.Where("action = ?", q.action)
	}
	if q.operatorID != 0 {
		tx = tx.Where("operator_id = ?", q.operatorID)
	}
	if q.keyword != "" {
		like := "%" + q.keyword + "%"
		tx = tx.Where("target_name ILIKE ? OR summary ILIKE ? OR operator_name ILIKE ?", like, like, like)
	}
	return tx
}

// ListLogs 操作日志列表（管理员）：按时间倒序分页。
// 刻意不返回 snapshot：单条快照可能很大（彻底删除项目会带上全部任务），
// 需要看快照时再调 GetLog 取单条。
func ListLogs(c *gin.Context) {
	q := parseLogQuery(c)

	var total int64
	logScope(q).Count(&total)

	var items []models.OperationLog
	logScope(q).Omit("snapshot").Order("id desc").
		Offset((q.page - 1) * q.pageSize).Limit(q.pageSize).Find(&items)

	ok(c, gin.H{
		"items":    items,
		"total":    total,
		"page":     q.page,
		"pageSize": q.pageSize,
	})
}

// GetLog 操作日志详情（含数据快照）
func GetLog(c *gin.Context) {
	id, good := parseUUID(c, "id")
	if !good {
		return
	}
	var row models.OperationLog
	if err := db.DB.First(&row, id).Error; err != nil {
		fail(c, http.StatusNotFound, "日志不存在")
		return
	}
	ok(c, row)
}
