package i18n

import (
	"sort"
	"strings"
)

// enMessages 英文文案目录：键就是源码里写死的中文原文。
//
// 好处是调用方零改动（fail / badRequest / okMessage 内部统一查表），
// 代价是「改了中文原文就会掉出目录」—— 所以下面每条都对着源码抄，
// 并且 i18n_test.go 会校验 handlers 里出现的中文串都能在目录里找到译文，
// 漏翻会在 go test 里直接暴露。
var enMessages = map[string]string{
	// 登录 / 通用
	"邮箱和密码不能为空": "Email and password are required",
	"邮箱或密码错误":   "Incorrect email or password",
	"令牌生成失败":    "Failed to generate the token",
	"邮箱格式不正确":   "Invalid email address",
	"参数错误":      "Invalid request",
	"保存失败":      "Failed to save",
	"创建失败":      "Failed to create",
	"删除失败":      "Failed to delete",
	"密码加密失败":    "Failed to hash the password",
	"保存附件失败":    "Failed to save the attachments",

	// 用户 / 部门
	"邮箱、用户名、密码均为必填":   "Email, username and password are required",
	"创建失败（邮箱可能已存在）":   "Failed to create (the email may already be in use)",
	"保存失败（邮箱可能已存在）":   "Failed to save (the email may already be in use)",
	"部门 ID 非法":        "Invalid department ID",
	"部门不存在":           "Department does not exist",
	"用户不存在":           "User does not exist",
	"部门名称不能为空":        "Department name is required",
	"部门名称已存在":         "The department name is already in use",
	"创建失败（部门名称可能已存在）": "Failed to create (the name may already be in use)",

	// 项目
	"项目不存在":          "Project does not exist",
	"无权访问该项目":        "You do not have access to this project",
	"无权查看该项目":        "You are not allowed to view this project",
	"标题不能为空":         "Title is required",
	"只有项目创建者或管理员可编辑": "Only the project creator or an administrator can edit it",
	"只有项目创建者或管理员可删除": "Only the project creator or an administrator can delete it",

	// 项目资料
	"资料不存在": "Doc does not exist",
	"只有项目成员可以新增项目资料":             "Only project members can add docs",
	"只有项目成员中的资料创建人、项目创建人或管理员可编辑": "Only the doc creator among project members, the project creator or an administrator can edit it",
	"只有项目成员中的资料创建人、项目创建人或管理员可删除": "Only the doc creator among project members, the project creator or an administrator can delete it",

	// 任务
	"只有项目成员可以创建任务": "Only project members can create tasks",
	"附件地址非法":       "Invalid attachment URL",
	"任务不存在":        "Task does not exist",
	"无权查看该任务":      "You are not allowed to view this task",
	"只有任务创建人、被指派人、项目创建人或管理员可编辑": "Only the task creator, assignee, project creator or an administrator can edit it",
	"状态非法": "Invalid status",
	"只有任务创建人、被指派人、项目创建人或管理员可移动任务": "Only the task creator, assignee, project creator or an administrator can move this task",
	"移动失败": "Failed to move the task",
	"只有管理员、任务创建人、被指派人或项目创建人可删除": "Only an administrator, the task creator, the assignee or the project creator can delete it",

	// 回收箱
	"该项目未被删除":          "This project is not deleted",
	"该任务未被删除":          "This task is not deleted",
	"该资料未被删除":          "This doc is not deleted",
	"只有项目创建者或管理员可恢复":   "Only the project creator or an administrator can restore it",
	"只有项目创建者或管理员可彻底删除": "Only the project creator or an administrator can purge it",
	"只有管理员、任务创建人、被指派人或项目创建人可恢复":    "Only an administrator, the task creator, the assignee or the project creator can restore it",
	"只有管理员、任务创建人、被指派人或项目创建人可彻底删除":  "Only an administrator, the task creator, the assignee or the project creator can purge it",
	"只有项目成员中的资料创建人、项目创建人或管理员可恢复":   "Only the doc creator among project members, the project creator or an administrator can restore it",
	"只有项目成员中的资料创建人、项目创建人或管理员可彻底删除": "Only the doc creator among project members, the project creator or an administrator can purge it",
	"任务所属项目已被删除，请先在回收箱恢复该项目":       "The project of this task is deleted; restore the project from the trash first",
	"资料所属项目已被删除，请先在回收箱恢复该项目":       "The project of this doc is deleted; restore the project from the trash first",

	// 附件上传 / 下载
	"未登录":      "Not signed in",
	"签发下载凭证失败": "Failed to issue the download ticket",
	"上传内容解析失败（可能超过大小限制）": "Failed to parse the upload (it may exceed the size limit)",
	"没有收到文件":   "No files received",
	"读取上传内容失败": "Failed to read the uploaded content",
	"登记上传文件失败": "Failed to register the uploaded file",

	// ID 解析 / 日志
	"ID 非法":    "Invalid ID",
	"用户 ID 非法": "Invalid user ID",
	"日志不存在":    "Log entry does not exist",

	// 鉴权中间件
	"登录已失效":   "Your session has expired",
	"缺少下载凭证":  "Missing download ticket",
	"下载凭证无效":  "Invalid download ticket",
	"需要管理员权限": "Administrator permission required",

	// 静态站点（web-dist 目录里的前端产物，见 server/webstatic.go）
	"接口不存在":     "Endpoint does not exist",
	"前端静态资源未构建": "Frontend assets are not built into this binary",
	"静态资源加载失败":  "Failed to load frontend assets",

	// 成功提示（okMessage）
	"已删除":   "Deleted",
	"已恢复":   "Restored",
	"已彻底删除": "Permanently deleted",
	"已登出":   "Signed out",
}

// enPrefixes 前缀型消息：原文 = 前缀 + 动态内容（错误详情、数量、文件名）。
var enPrefixes = map[string]string{
	"保存失败: ":    "Failed to save: ",
	"恢复失败：":     "Restore failed: ",
	"彻底删除失败：":   "Purge failed: ",
	"关联成员失败：":   "Failed to link the members: ",
	"保存成员失败：":   "Failed to save the members: ",
	"一次最多上传 ":   "Up to ",
	"该部门下还有 ":   "This department still has ",
	"附件「":       "Attachment \"",
	"不支持的文件类型 ": "Unsupported file type ",
	"无法识别文件类型：": "Unable to determine the file type: ",
	"文件超过大小上限 ": "The file exceeds the limit of ",
}

// enSuffixes 尾缀型消息：与 enPrefixes 组合成「前缀 + 动态内容 + 尾缀」，
// 或者配合「」包裹的动态内容（此时前缀以引号收尾）。
var enSuffixes = map[string]string{
	" 个文件": " files at a time",
	" 个用户，请先调整归属后再删除": " users; reassign them before deleting",
	"不存在":    "\" does not exist",
	"超过大小上限": "\" exceeds the size limit",
	" MB":    " MB",
}

// translate 中文原文 → 英文。中文与未收录的原文原样返回。
func translate(lang Lang, msg string) string {
	if lang != EN || msg == "" {
		return msg
	}
	// 1) 整句命中
	if to, ok := enMessages[msg]; ok {
		return to
	}

	// 2) 前缀命中（按前缀长度从长到短，保证结果是确定的）
	head, rest, ok := cutPrefix(msg, enPrefixes)
	if ok {
		// 2a) 前缀 + 「动态内容」 + 尾缀：附件「a.png」不存在
		if inner, tail, found := strings.Cut(rest, "」"); found {
			return head + inner + translateTail(tail)
		}
		// 2b) 前缀 + 动态内容 + 尾缀：一次最多上传 3 个文件
		if mid, to, found := cutSuffix(rest, enSuffixes); found {
			return head + mid + to
		}
		// 2c) 前缀 + 动态内容：保存失败: <数据库错误>
		return head + rest
	}

	// 3) 「动态内容」 + 尾缀：「a.png」超过大小上限
	if inner, ok := strings.CutPrefix(msg, "「"); ok {
		if name, tail, found := strings.Cut(inner, "」"); found {
			return "\"" + name + translateTail(tail)
		}
	}
	return msg
}

// translateTail 翻译「」之后的尾缀；未收录时原样返回。
func translateTail(tail string) string {
	if to, ok := enSuffixes[tail]; ok {
		return to
	}
	return tail
}

// cutPrefix 按最长匹配切出前缀，返回 (前缀译文, 剩余, 是否命中)。
func cutPrefix(msg string, table map[string]string) (string, string, bool) {
	for _, k := range sortedKeys(table) {
		if strings.HasPrefix(msg, k) {
			return table[k], strings.TrimPrefix(msg, k), true
		}
	}
	return "", msg, false
}

// cutSuffix 按最长匹配切出尾缀，返回 (剩余, 尾缀译文, 是否命中)。
func cutSuffix(msg string, table map[string]string) (string, string, bool) {
	for _, k := range sortedKeys(table) {
		if strings.HasSuffix(msg, k) {
			return strings.TrimSuffix(msg, k), table[k], true
		}
	}
	return msg, "", false
}

// sortedKeys 按长度从长到短返回键：多个前缀可能互相包含时，先试长的。
func sortedKeys(table map[string]string) []string {
	keys := make([]string, 0, len(table))
	for k := range table {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	return keys
}
