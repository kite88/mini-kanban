// Package storage 负责附件落盘与读取。
// 目录与大小上限来自 config.storage（dir / maxSizeMB），不做代码兜底。
package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"mini-kanban/internal/config"
)

var (
	dir     string
	maxSize int64
)

// MIMEByExt 允许上传的扩展名 -> 返回给浏览器的 Content-Type。
// 不在表内的类型一律拒绝（含 .svg：SVG 可携带脚本，不作为可上传类型）。
var MIMEByExt = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".ico":  "image/x-icon",
	".pdf":  "application/pdf",
	".txt":  "text/plain; charset=utf-8",
	".md":   "text/markdown; charset=utf-8",
	".csv":  "text/csv; charset=utf-8",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".zip":  "application/zip",
	".7z":   "application/x-7z-compressed",
	".rar":  "application/vnd.rar",
	".gz":   "application/gzip",
	".mp4":  "video/mp4",
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
}

// storedNameRe 落盘文件名规则：32 位随机 hex + 白名单扩展名，由服务端生成，
// 用户提供的文件名永远不进入路径（防目录穿越）。
var storedNameRe = regexp.MustCompile(`^[0-9a-f]{32}\.[a-z0-9]{1,8}$`)

// URLPrefix 附件对外地址前缀（相对 API 根；前端按 API 基址补全）
const URLPrefix = "/api/files/"

// Init 初始化存储目录（不存在则创建）。
func Init(cfg *config.Config) error {
	dir = filepath.Clean(strings.TrimSpace(cfg.Storage.Dir))
	maxSize = int64(cfg.Storage.MaxSizeMB) << 20
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建存储目录 %s 失败：%w", dir, err)
	}
	return nil
}

// Dir 存储目录
func Dir() string { return dir }

// MaxSize 单文件大小上限（字节）
func MaxSize() int64 { return maxSize }

// Allowed 该扩展名是否允许上传
func Allowed(name string) bool {
	_, ok := MIMEByExt[strings.ToLower(filepath.Ext(name))]
	return ok
}

// MIME 按落盘文件名返回 Content-Type
func MIME(name string) (string, bool) {
	ct, ok := MIMEByExt[strings.ToLower(filepath.Ext(name))]
	return ct, ok
}

// URL 由落盘文件名得到对外相对地址
func URL(name string) string { return URLPrefix + name }

// NameFromURL 解析并校验 /api/files/<随机名> 形式的地址，返回落盘文件名
func NameFromURL(u string) (string, bool) {
	name := strings.TrimPrefix(strings.TrimSpace(u), URLPrefix)
	if name == strings.TrimSpace(u) || !storedNameRe.MatchString(name) {
		return "", false
	}
	return name, true
}

// Save 把上传内容写入存储目录，返回落盘文件名与写入字节数。
// 超过大小上限会立即失败并清理半截文件。
func Save(originalName string, r io.Reader) (string, int64, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	if _, ok := MIMEByExt[ext]; !ok {
		if ext == "" {
			return "", 0, fmt.Errorf("无法识别文件类型：%s", originalName)
		}
		return "", 0, fmt.Errorf("不支持的文件类型 %s", ext)
	}

	// 16 字节随机 → 32 位 hex，与 storedNameRe / 富文本里的图片地址规则保持一致
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", 0, fmt.Errorf("生成文件名失败：%w", err)
	}
	name := hex.EncodeToString(raw) + ext
	path := filepath.Join(dir, name)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", 0, fmt.Errorf("写入文件失败：%w", err)
	}
	// 多读 1 字节用于判断是否超限
	n, err := io.Copy(f, io.LimitReader(r, maxSize+1))
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		os.Remove(path)
		if err == nil {
			err = closeErr
		}
		return "", 0, fmt.Errorf("写入文件失败：%w", err)
	}
	if n > maxSize {
		os.Remove(path)
		return "", 0, fmt.Errorf("文件超过大小上限 %d MB", maxSize>>20)
	}
	return name, n, nil
}

// Path 校验落盘文件名并返回绝对路径与文件信息（仅接受服务端生成的随机名）。
func Path(name string) (string, os.FileInfo, bool) {
	if !storedNameRe.MatchString(name) {
		return "", nil, false
	}
	p := filepath.Join(dir, name)
	st, err := os.Stat(p)
	if err != nil || st.IsDir() {
		return "", nil, false
	}
	return p, st, true
}
