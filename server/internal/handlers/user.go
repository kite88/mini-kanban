package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"mini-kanban/internal/db"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/models"
	"mini-kanban/internal/oplog"
	"mini-kanban/internal/utils"
)

type userInput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	// DepartmentID nil 表示不变更；空串表示清空（未分配）；否则是部门 UUID
	DepartmentID *string     `json:"departmentId"`
	Password     string      `json:"password"`
	Role         models.Role `json:"role"`
}

// departmentText 部门的展示文本：未分配时给出明确文案，避免日志摘要里出现空白
func departmentText(id *models.UUID) string {
	if name := oplog.DepartmentName(id); name != "" {
		return name
	}
	return "未分配"
}

// setDepartment 按入参设置用户部门：nil 保持不变、空串清空（写 NULL）、否则指定部门。
// 部门 id 是 UUID，没有「0 清空」这种整型哨兵，空串才代表「未分配」。
// 返回错误提示与是否通过。
func setDepartment(u *models.User, departmentID *string) (string, bool) {
	if departmentID == nil {
		return "", true
	}
	raw := strings.TrimSpace(*departmentID)
	if raw == "" {
		u.DepartmentID = nil
		return "", true
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return "部门 ID 非法", false
	}
	if !departmentExists(id) {
		return "部门不存在", false
	}
	u.DepartmentID = &id
	return "", true
}

func ListUsers(c *gin.Context) {
	var users []models.User
	db.DB.Order("id asc").Find(&users)
	ok(c, users)
}

func CreateUser(c *gin.Context) {
	var req userInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	req.Email = utils.NormalizeEmail(req.Email)
	if req.Email == "" || req.Username == "" || req.Password == "" {
		badRequest(c, "邮箱、用户名、密码均为必填")
		return
	}
	if !utils.IsEmail(req.Email) {
		badRequest(c, "邮箱格式不正确")
		return
	}
	if req.Role == "" {
		req.Role = models.RoleMember
	}
	u := models.User{Email: req.Email, Username: req.Username, Role: req.Role}
	if msg, pass := setDepartment(&u, req.DepartmentID); !pass {
		badRequest(c, msg)
		return
	}
	if err := u.SetPassword(req.Password); err != nil {
		fail(c, http.StatusInternalServerError, "密码加密失败")
		return
	}
	if err := db.DB.Create(&u).Error; err != nil {
		badRequest(c, "创建失败（邮箱可能已存在）")
		return
	}
	invalidate("users")
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionCreate,
		TargetType: models.TargetUser,
		TargetID:   u.ID.String(),
		TargetName: u.Username,
		Summary: "邮箱 " + u.Email + "；角色 " + models.RoleLabel(u.Role) +
			"；部门 " + departmentText(u.DepartmentID),
	})
	created(c, u)
}

func UpdateUser(c *gin.Context) {
	id, good := parseUserID(c, "id")
	if !good {
		return
	}
	var u models.User
	if err := db.DB.First(&u, id).Error; err != nil {
		fail(c, http.StatusNotFound, "用户不存在")
		return
	}
	var req userInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	// 保留修改前的内容，供操作日志对比（值拷贝，互不影响）
	before := u
	if email := utils.NormalizeEmail(req.Email); email != "" {
		if !utils.IsEmail(email) {
			badRequest(c, "邮箱格式不正确")
			return
		}
		u.Email = email
	}
	if req.Username != "" {
		u.Username = req.Username
	}
	if msg, pass := setDepartment(&u, req.DepartmentID); !pass {
		badRequest(c, msg)
		return
	}
	if req.Role != "" {
		u.Role = req.Role
	}
	if req.Password != "" {
		_ = u.SetPassword(req.Password)
	}
	if err := db.DB.Save(&u).Error; err != nil {
		badRequest(c, "保存失败（邮箱可能已存在）")
		return
	}
	invalidate("users", "projects", "tasks")
	summary := oplog.ChangesSummary(
		oplog.Change{Label: "用户名", From: before.Username, To: u.Username},
		oplog.Change{Label: "邮箱", From: before.Email, To: u.Email},
		oplog.Change{Label: "部门", From: departmentText(before.DepartmentID), To: departmentText(u.DepartmentID)},
		oplog.Change{Label: "角色", From: models.RoleLabel(before.Role), To: models.RoleLabel(u.Role)},
	)
	if req.Password != "" {
		summary = oplog.AppendSummary(summary, "密码已重置")
	}
	oplog.Record(middleware.CurrentUser(c), oplog.Entry{
		Action:     models.ActionUpdate,
		TargetType: models.TargetUser,
		TargetID:   u.ID.String(),
		TargetName: u.Username,
		Summary:    summary,
	})
	ok(c, u)
}

func DeleteUser(c *gin.Context) {
	id, good := parseUserID(c, "id")
	if !good {
		return
	}
	var u models.User
	if err := db.DB.First(&u, id).Error; err != nil {
		fail(c, http.StatusNotFound, "用户不存在")
		return
	}
	op := middleware.CurrentUser(c)
	// 删除前采集数据快照：用户是物理删除，删掉之后只剩这份留档
	snapshot := oplog.UserSnapshot(u.ID)

	// 删除动作与快照写在同一个事务里，避免「数据删了、快照却没留下」
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// 被指派的任务解除关联，避免外键约束（fk_tasks_assignee）导致删除失败。
		// 必须带 Unscoped：软删除的历史任务同样持有 assignee_id，会被外键拦住。
		if err := tx.Unscoped().Model(&models.Task{}).Where("assignee_id = ?", id).
			Update("assignee_id", nil).Error; err != nil {
			return err
		}
		// 已软删除的项目仍持有 owner_id（fk_projects_owner），owner_id 又不能置空，
		// 这里把归属转给执行删除的管理员：既让删除得以进行，也保住软删除数据的可恢复性。
		// （未删除的项目归属不动：该用户仍是项目创建者时依旧会因外键删除失败，这是原有行为。）
		if op != nil {
			if err := tx.Unscoped().Model(&models.Project{}).
				Where("owner_id = ? AND deleted_at IS NOT NULL", id).
				Update("owner_id", op.ID).Error; err != nil {
				return err
			}
		}
		// 项目成员关联一并清理
		if err := tx.Where("user_id = ?", id).Delete(&models.ProjectMember{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.User{}, id).Error; err != nil {
			return err
		}
		return oplog.RecordTx(tx, op, oplog.Entry{
			Action:     models.ActionDelete,
			TargetType: models.TargetUser,
			TargetID:   u.ID.String(),
			TargetName: u.Username,
			Summary: "物理删除，不可恢复；邮箱 " + u.Email +
				"；角色 " + models.RoleLabel(u.Role) + "；部门 " + departmentText(u.DepartmentID),
			Snapshot: snapshot,
		})
	})
	if err != nil {
		badRequest(c, "删除失败")
		return
	}
	invalidate("users", "projects", "tasks")
	okMessage(c, "已删除")
}
