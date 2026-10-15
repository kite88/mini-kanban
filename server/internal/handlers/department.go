package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mini-kanban/internal/db"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/models"
	"mini-kanban/internal/oplog"
)

type departmentInput struct {
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

// departmentItem 部门列表项：附带人数，便于前端展示与删除前提示
type departmentItem struct {
	models.Department
	UserCount int64 `json:"userCount"`
}

// departmentID 部门 id（UUID v7）
func departmentID(c *gin.Context) (models.UUID, bool) {
	return parseUUID(c, "id")
}

// departmentNameExists 判断部门名是否已被占用（excludeID 用于编辑时排除自身）
func departmentNameExists(name string, excludeID models.UUID) bool {
	q := db.DB.Model(&models.Department{}).Where("name = ?", name)
	if excludeID != (models.UUID{}) {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	q.Count(&n)
	return n > 0
}

// departmentExists 判断部门是否存在（供用户保存时校验 departmentId）
func departmentExists(id models.UUID) bool {
	var n int64
	db.DB.Model(&models.Department{}).Where("id = ?", id).Count(&n)
	return n > 0
}

func ListDepartments(c *gin.Context) {
	var list []models.Department
	db.DB.Order("id asc").Find(&list)

	res := make([]departmentItem, 0, len(list))
	for _, d := range list {
		var n int64
		db.DB.Model(&models.User{}).Where("department_id = ?", d.ID).Count(&n)
		res = append(res, departmentItem{Department: d, UserCount: n})
	}
	ok(c, res)
}

func GetDepartment(c *gin.Context) {
	id, good := departmentID(c)
	if !good {
		return
	}
	var d models.Department
	if err := db.DB.First(&d, id).Error; err != nil {
		fail(c, http.StatusNotFound, "部门不存在")
		return
	}
	var n int64
	db.DB.Model(&models.User{}).Where("department_id = ?", id).Count(&n)
	ok(c, departmentItem{Department: d, UserCount: n})
}

func CreateDepartment(c *gin.Context) {
	var req departmentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		badRequest(c, "部门名称不能为空")
		return
	}
	if departmentNameExists(name, models.UUID{}) {
		badRequest(c, "部门名称已存在")
		return
	}
	d := models.Department{Name: name, Remark: strings.TrimSpace(req.Remark)}
	if err := db.DB.Create(&d).Error; err != nil {
		badRequest(c, "创建失败（部门名称可能已存在）")
		return
	}
	invalidate("departments", "users")
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionCreate,
		TargetType: models.TargetDepartment,
		TargetID:   d.ID.String(),
		TargetName: d.Name,
		Summary:    oplog.ChangesSummary(oplog.Change{Label: "备注", To: d.Remark}),
	})
	created(c, departmentItem{Department: d})
}

func UpdateDepartment(c *gin.Context) {
	id, good := departmentID(c)
	if !good {
		return
	}
	var d models.Department
	if err := db.DB.First(&d, id).Error; err != nil {
		fail(c, http.StatusNotFound, "部门不存在")
		return
	}
	var req departmentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		badRequest(c, "部门名称不能为空")
		return
	}
	if departmentNameExists(name, id) {
		badRequest(c, "部门名称已存在")
		return
	}
	// 保留修改前的内容，供操作日志对比（值拷贝，互不影响）
	before := d
	d.Name = name
	d.Remark = strings.TrimSpace(req.Remark)
	if err := db.DB.Save(&d).Error; err != nil {
		fail(c, http.StatusInternalServerError, "保存失败")
		return
	}
	var n int64
	db.DB.Model(&models.User{}).Where("department_id = ?", id).Count(&n)
	invalidate("departments", "users")
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionUpdate,
		TargetType: models.TargetDepartment,
		TargetID:   d.ID.String(),
		TargetName: d.Name,
		Summary: oplog.ChangesSummary(
			oplog.Change{Label: "部门名称", From: before.Name, To: d.Name},
			oplog.Change{Label: "备注", From: before.Remark, To: d.Remark},
		),
	})
	ok(c, departmentItem{Department: d, UserCount: n})
}

func DeleteDepartment(c *gin.Context) {
	id, good := departmentID(c)
	if !good {
		return
	}
	var d models.Department
	if err := db.DB.First(&d, id).Error; err != nil {
		fail(c, http.StatusNotFound, "部门不存在")
		return
	}
	// 有成员时不允许删除，避免用户指向一个不存在的部门
	var n int64
	db.DB.Model(&models.User{}).Where("department_id = ?", id).Count(&n)
	if n > 0 {
		badRequest(c, "该部门下还有 "+strconv.FormatInt(n, 10)+" 个用户，请先调整归属后再删除")
		return
	}
	op := middleware.CurrentUser(c)
	// 删除前采集数据快照：部门是物理删除，删掉之后只剩这份留档
	snapshot := oplog.DepartmentSnapshot(d.ID)
	// 删除动作与快照写在同一个事务里，避免「数据删了、快照却没留下」
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Department{}, d.ID).Error; err != nil {
			return err
		}
		return oplog.RecordTx(tx, op, oplog.Entry{
			Action:     models.ActionDelete,
			TargetType: models.TargetDepartment,
			TargetID:   d.ID.String(),
			TargetName: d.Name,
			Summary:    "物理删除，不可恢复；删除时该部门下已无成员",
			Snapshot:   snapshot,
		})
	})
	if err != nil {
		fail(c, http.StatusInternalServerError, "删除失败")
		return
	}
	invalidate("departments", "users")
	okMessage(c, "已删除")
}
