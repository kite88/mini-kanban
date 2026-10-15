package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mini-kanban/internal/db"
	"mini-kanban/internal/middleware"
	"mini-kanban/internal/models"
	"mini-kanban/internal/storage"
	"mini-kanban/internal/utils"
)

// maxUploadFiles 单次上传的文件个数上限
const maxUploadFiles = 10

// fileTicketTTL 附件下载票据有效期：短时效。
// 下载 URL 里只能出现这张票据（不会出现登录 token），即使被转发 / 记录，
// 攻击者拿到也只是「一段时间内能读附件」，不能调用其他接口。
const fileTicketTTL = 2 * time.Hour

// IssueFileTicket 签发附件下载票据（需登录态，只认 Authorization 头）。
// 前端拿到后拼成 /api/files/xxx?ticket=... 用于 <img src> / 附件链接。
func IssueFileTicket(c *gin.Context) {
	u := middleware.CurrentUser(c)
	if u == nil {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	var sessionJTI string
	if v, exists := c.Get(middleware.CtxClaims); exists {
		if claims, ok := v.(*utils.Claims); ok {
			sessionJTI = claims.JTI
		}
	}
	ticket, exp, err := utils.GenerateFileTicket(u, sessionJTI, fileTicketTTL)
	if err != nil {
		fail(c, http.StatusInternalServerError, "签发下载凭证失败")
		return
	}
	ok(c, gin.H{
		"ticket":    ticket,
		"expiresAt": exp,
		"expiresIn": int(time.Until(exp).Seconds()),
	})
}

// UploadFiles 接收 multipart/form-data（字段名 files，可重复）并落盘，
// 返回 [{name, url, size}]：name 是原始文件名（仅展示），url 是相对地址。
func UploadFiles(c *gin.Context) {
	// 整个请求体上限：单文件上限 × 个数上限
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, storage.MaxSize()*maxUploadFiles)

	form, err := c.MultipartForm()
	if err != nil {
		badRequest(c, "上传内容解析失败（可能超过大小限制）")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		badRequest(c, "没有收到文件")
		return
	}
	if len(files) > maxUploadFiles {
		badRequest(c, "一次最多上传 "+strconv.Itoa(maxUploadFiles)+" 个文件")
		return
	}

	type uploaded struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Size int64  `json:"size"`
	}
	u := middleware.CurrentUser(c)
	out := make([]uploaded, 0, len(files))
	for _, fh := range files {
		if fh.Size > storage.MaxSize() {
			fail(c, http.StatusRequestEntityTooLarge, "「"+fh.Filename+"」超过大小上限")
			return
		}
		src, err := fh.Open()
		if err != nil {
			fail(c, http.StatusBadRequest, "读取上传内容失败")
			return
		}
		name, size, err := storage.Save(fh.Filename, src)
		src.Close()
		if err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		// 登记上传者：文件还没被任何数据引用时（编辑器预览 / 保存前），
		// 下载鉴权只认这一条记录。登记失败就把刚落盘的文件删掉，
		// 避免留下「磁盘上有、却谁都判不了归属」的文件。
		if err := db.DB.Create(&models.UploadedFile{Name: name, UploaderID: u.ID}).Error; err != nil {
			if path, _, ok := storage.Path(name); ok {
				os.Remove(path)
			}
			fail(c, http.StatusInternalServerError, "登记上传文件失败")
			return
		}
		out = append(out, uploaded{Name: CleanFileName(fh.Filename), URL: storage.URL(name), Size: size})
	}
	ok(c, out)
}

// ServeFile 返回上传的文件（需登录态：?ticket=<下载票据>，且对文件所属数据有权限）。
// 富文本里的 <img src> 与附件链接由浏览器直接发起，带不了 Authorization 头，
// 所以凭据只能放在 URL 里；为不泄露登录凭据，这里只接受 middleware.FileTicketRequired
// 认可的下载票据（见 IssueFileTicket）。
//
// 票据只说明「谁在下载」，还要再判「他能不能下这个文件」：
// 鉴权通过但对该文件的数据没有权限时，同样返回 404（不泄露文件是否存在）。
func ServeFile(c *gin.Context) {
	name := c.Param("name")
	path, info, good := storage.Path(name)
	if !good {
		c.Status(http.StatusNotFound)
		return
	}
	if !canDownloadFile(middleware.CurrentUser(c), name) {
		c.Status(http.StatusNotFound)
		return
	}
	contentType, allowed := storage.MIME(name)
	if !allowed {
		c.Status(http.StatusNotFound)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer f.Close()

	c.Header("Content-Type", contentType)
	// 禁止按内容嗅探类型，避免上传内容被当成 HTML / JS 执行
	c.Header("X-Content-Type-Options", "nosniff")
	// 响应依赖登录态：private 避免被代理 / CDN 缓存后绕过鉴权，
	// 缓存时长也不宜过长，否则登出后浏览器仍可能直接命中旧响应。
	c.Header("Cache-Control", "private, max-age=3600")
	c.Header("Vary", "Authorization")
	if !strings.HasPrefix(contentType, "image/") && !strings.HasPrefix(contentType, "video/") {
		// 非图片 / 视频一律作为下载，并尽量还原原始文件名
		c.Header("Content-Disposition", `attachment; filename="`+downloadName(name)+`"`)
	}
	http.ServeContent(c.Writer, c.Request, name, info.ModTime(), f)
}

// canDownloadFile 判断当前用户能否下载某个落盘文件。
//
// 附件表里只有 url、不带权限信息，所以这里按「谁引用了它」反查归属，再套用对应入口的可见性
// （原则是「即看即下」：能在界面上看到这条数据，就能下它的附件）：
//  1. 资料附件 / 资料正文里的内嵌图片 → 按资料所属项目的可见性
//     （与资料弹窗一致：能看到项目就能看资料）；
//  2. 任务附件 / 任务描述里的内嵌图片 → 按任务可见性
//     （与看板上能看到该任务的人一致：管理员、任务创建人、被指派人、项目创建人）；
//  3. 只剩「已删除的引用」（附件行已软删，或主体已软删但曾引用过正文）→ 拒绝：
//     数据都删了，文件也不该再被下载；
//  4. 没有任何引用记录的文件，按上传登记（uploaded_files）判定：
//     曾被引用过但数据已删除 → 拒绝；从未被引用（编辑器刚粘贴、上传后还没保存）
//     → 只允许上传者本人，既不挡住自己的预览，也不会变成「谁有地址谁能下」的孤儿。
func canDownloadFile(u *models.User, stored string) bool {
	if u == nil {
		return false
	}
	url := storage.URL(stored)

	// 1) 资料附件：按资料所属项目的可见性
	// 用切片接 Pluck：UUID 是自定义类型，标量目标不保证被 gorm 正确扫描
	// （会把整段 UUID 字符串拆成字节逐个赋值，扫描直接报错且结果停在零值，
	//  表现为「附件行明明存在、归属却查不出来」——见 project.go 的同类写法）
	var docIDs []models.UUID
	if err := db.DB.Model(&models.ProjectDocAttachment{}).Where("url = ?", url).
		Order("id asc").Limit(1).Pluck("doc_id", &docIDs).Error; err == nil && len(docIDs) > 0 {
		var doc models.ProjectDoc
		// 附件行还在、资料却查不到（已软删）时直接拒绝，不落到下面的放行分支。
		// 用 Find 而不是 First：查不到是这里的正常分支，不必往服务日志里写 record not found
		if err := db.DB.Limit(1).Find(&doc, docIDs[0]).Error; err != nil || doc.ID == uuid.Nil {
			return false
		}
		return canViewProjectByID(u, doc.ProjectID)
	}

	// 2) 任务附件：按任务可见性（即看即下：看得见这个任务的人才能下它的附件）
	// 同上：UUID 列的 Pluck 必须用切片接
	var taskIDs []models.UUID
	if err := db.DB.Model(&models.TaskAttachment{}).Where("url = ?", url).
		Order("id asc").Limit(1).Pluck("task_id", &taskIDs).Error; err == nil && len(taskIDs) > 0 {
		var t models.Task
		if err := db.DB.Limit(1).Find(&t, taskIDs[0]).Error; err != nil || t.ID == uuid.Nil {
			return false
		}
		return canManageTask(u, &t)
	}

	// 3) 富文本内嵌图片：查未删除的正文引用，按正文所属数据的口径判定
	var doc models.ProjectDoc
	if err := db.DB.Where("content LIKE ?", "%"+url+"%").Limit(1).Find(&doc).Error; err == nil && doc.ID != uuid.Nil {
		return canViewProjectByID(u, doc.ProjectID)
	}
	var t models.Task
	if err := db.DB.Where("description LIKE ?", "%"+url+"%").Limit(1).Find(&t).Error; err == nil && t.ID != uuid.Nil {
		return canManageTask(u, &t)
	}

	// 4) 没有任何未删引用：曾被引用过（哪怕现在已删除）就拒绝；
	//    否则只允许上传者本人（编辑器预览 / 上传后还没保存）
	if referencedFile(url) {
		return false
	}
	var f models.UploadedFile
	if err := db.DB.Where("name = ?", stored).Limit(1).Find(&f).Error; err != nil || f.ID == uuid.Nil {
		return false
	}
	return f.UploaderID == u.ID
}

// canViewProjectByID 按项目 id 判定可见性（下载校验用）
func canViewProjectByID(u *models.User, id models.UUID) bool {
	var p models.Project
	if err := db.DB.Limit(1).Find(&p, id).Error; err != nil || p.ID == uuid.Nil {
		return false
	}
	return canViewProject(u, &p)
}

// referencedFile 该文件是否曾被任何数据引用过（含已软删除的资料 / 任务与附件行）
func referencedFile(url string) bool {
	var n int64
	db.DB.Unscoped().Model(&models.ProjectDocAttachment{}).Where("url = ?", url).Count(&n)
	if n > 0 {
		return true
	}
	db.DB.Unscoped().Model(&models.TaskAttachment{}).Where("url = ?", url).Count(&n)
	if n > 0 {
		return true
	}
	db.DB.Unscoped().Model(&models.ProjectDoc{}).Where("content LIKE ?", "%"+url+"%").Count(&n)
	if n > 0 {
		return true
	}
	db.DB.Unscoped().Model(&models.Task{}).Where("description LIKE ?", "%"+url+"%").Count(&n)
	return n > 0
}

// downloadName 下载时展示的文件名：优先用附件表里的原始名，取不到则用落盘名。
func downloadName(stored string) string {
	url := storage.URL(stored)
	var a models.TaskAttachment
	if err := db.DB.Where("url = ?", url).Limit(1).Find(&a).Error; err == nil && a.ID != uuid.Nil {
		if n := CleanFileName(a.Name); n != "" {
			return n
		}
	}
	// 项目资料附件也要还原成上传时的原始文件名
	var d models.ProjectDocAttachment
	if err := db.DB.Where("url = ?", url).Limit(1).Find(&d).Error; err == nil && d.ID != uuid.Nil {
		if n := CleanFileName(d.Name); n != "" {
			return n
		}
	}
	return stored
}

// CleanFileName 清理原始文件名：去掉目录部分与控制字符，限制长度。
func CleanFileName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, strings.TrimSpace(name))
	if name == "" || name == "." || name == ".." {
		return "未命名文件"
	}
	if rs := []rune(name); len(rs) > 120 {
		name = string(rs[:120])
	}
	return name
}
