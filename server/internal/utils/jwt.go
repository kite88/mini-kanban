package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"mini-kanban/internal/config"
	"mini-kanban/internal/models"
)

var (
	cfg *config.Config
	// fileTicketKey 附件下载票据的独立签名密钥，由 jwt.secret 单向派生（见 Init）。
	// 密钥分离后，两个方向的凭据即使某一侧漏检 aud 也无法互换使用。
	fileTicketKey []byte
)

// fileTicketLabel 派生下载票据密钥的领域标签，避免与其它用途的派生值撞车
const fileTicketLabel = "mini-kanban/file-ticket/v1"

func Init(c *config.Config) {
	cfg = c
	// HMAC-SHA256 单向派生：拿到票据密钥也推不出登录密钥
	mac := hmac.New(sha256.New, []byte(c.JWT.Secret))
	mac.Write([]byte(fileTicketLabel))
	fileTicketKey = mac.Sum(nil)
}

type Claims struct {
	// UserID 用户 id（雪花 ID 的数值形态）：JWT 只在服务端解析，按数字存即可
	UserID           int64       `json:"uid"`
	Role             models.Role `json:"role"`
	JTI              string      `json:"jti"`
	RegisteredClaims jwt.RegisteredClaims
}

func (c Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetExpirationTime()
}
func (c Claims) GetIssuedAt() (*jwt.NumericDate, error)  { return c.RegisteredClaims.GetIssuedAt() }
func (c Claims) GetNotBefore() (*jwt.NumericDate, error) { return c.RegisteredClaims.GetNotBefore() }
func (c Claims) GetIssuer() (string, error)              { return c.RegisteredClaims.GetIssuer() }
func (c Claims) GetSubject() (string, error)             { return c.RegisteredClaims.GetSubject() }
func (c Claims) GetAudience() (jwt.ClaimStrings, error)  { return c.RegisteredClaims.GetAudience() }

// FileAudience 附件下载票据的受众标识：下载接口只接受 aud 含 file 的票据。
// 票据另有独立签名密钥（fileTicketKey），aud 校验属于第二道防线。
const FileAudience = "file"

// IsFileTicket 该票据是否为附件下载专用（aud 含 file）。
// 密钥已经分离，这里保留显式判断作为纵深防御，供 middleware.AuthRequired 拒绝票据。
func (c Claims) IsFileTicket() bool {
	for _, a := range c.RegisteredClaims.Audience {
		if a == FileAudience {
			return true
		}
	}
	return false
}

func GenerateToken(u *models.User) (string, *Claims, error) {
	jti := randomID()
	now := time.Now()
	claims := Claims{
		UserID: int64(u.ID),
		Role:   u.Role,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.Email,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.JWT.TTLHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWT.Secret))
	return signed, &claims, err
}

// GenerateFileTicket 签发附件下载票据：只具备读取附件的能力，不能调用其他接口。
// 下载地址由浏览器直接请求（<img src> / <a href>），只能把凭据放进 URL，
// 因此这里签一张短时效、低权限的票据替代登录 token，避免 URL 泄露等于登录态泄露。
// 用独立派生密钥签名（非 jwt.secret），JTI 沿用登录会话的 jti，
// 登出时该 jti 进黑名单，票据随之立即失效。
func GenerateFileTicket(u *models.User, sessionJTI string, ttl time.Duration) (string, time.Time, error) {
	if sessionJTI == "" {
		sessionJTI = randomID()
	}
	now := time.Now()
	exp := now.Add(ttl)
	claims := Claims{
		UserID: int64(u.ID),
		Role:   u.Role,
		JTI:    sessionJTI,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.Email,
			Audience:  jwt.ClaimStrings{FileAudience},
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(fileTicketKey)
	return signed, exp, err
}

// ParseToken 校验登录 token：固定 HS256、必须带 exp，用 jwt.secret 验签
func ParseToken(s string) (*Claims, error) {
	return parseJWT(s, []byte(cfg.JWT.Secret))
}

// ParseFileTicket 校验附件下载票据：用独立的派生密钥验签，并强制 aud 必须为 file。
// 密钥不同 → 登录 token 在这里一律验签失败；aud 校验 → 即使密钥相同也换不了用途。
func ParseFileTicket(s string) (*Claims, error) {
	return parseJWT(s, fileTicketKey, jwt.WithAudience(FileAudience))
}

// parseJWT 统一下发解析选项：只允许 HS256（防算法混淆），且必须带过期时间
func parseJWT(s string, key []byte, extra ...jwt.ParserOption) (*Claims, error) {
	opts := append([]jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	}, extra...)
	claims := &Claims{}
	if _, err := jwt.ParseWithClaims(s, claims, func(*jwt.Token) (interface{}, error) {
		return key, nil
	}, opts...); err != nil {
		return nil, err
	}
	return claims, nil
}

var ErrUnauthorized = errors.New("unauthorized")

func randomID() string {
	return time.Now().Format("20060102150405.000000000") + randSuffix()
}
