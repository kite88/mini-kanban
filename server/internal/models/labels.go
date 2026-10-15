package models

// 枚举值的中文名：只用于操作日志的摘要与快照展示，
// 不影响接口字段本身（接口仍下发 todo / admin 这类稳定值）。

// TaskStatusLabels 任务状态中文名（与前端看板列名一致）
var TaskStatusLabels = map[string]string{
	"todo":     "待办",
	"doing":    "进行中",
	"done":     "已完成",
	"accepted": "已验收",
}

// StatusLabel 取任务状态的中文名，未知状态原样返回
func StatusLabel(s string) string {
	if v, ok := TaskStatusLabels[s]; ok {
		return v
	}
	return s
}

// RoleLabels 角色中文名
var RoleLabels = map[Role]string{
	RoleAdmin:  "管理员",
	RoleMember: "成员",
}

// RoleLabel 取角色中文名，未知角色原样返回
func RoleLabel(r Role) string {
	if v, ok := RoleLabels[r]; ok {
		return v
	}
	return string(r)
}
