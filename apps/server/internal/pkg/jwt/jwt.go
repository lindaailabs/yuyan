// Package jwt 封装 token 签发与校验（HS256）。
// claims：uid（用户 id）、typ（access/refresh）、exp；密钥经环境变量 JWT_SECRET 注入，禁止硬编码（guide §9.3）。
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// TypeAccess 业务接口调用凭证，有效期 2h。
	TypeAccess = "access"
	// TypeRefresh 用于换发新 token，有效期 30d。
	TypeRefresh = "refresh"

	accessTTL  = 2 * time.Hour
	refreshTTL = 30 * 24 * time.Hour
)

// claims 自定义载荷。
type claims struct {
	UID int64  `json:"uid"`
	Typ string `json:"typ"`
	jwt.RegisteredClaims
}

// TokenPair 登录/刷新接口的返回。
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // access 有效期（秒）
}

// Manager 签发与解析器。
type Manager struct {
	secret []byte
}

// NewManager 构造（secret 为 JWT_SECRET 环境变量注入的密钥）。
func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

// Issue 签发双 token。
func (m *Manager) Issue(uid int64) (TokenPair, error) {
	access, err := m.sign(uid, TypeAccess, accessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := m.sign(uid, TypeRefresh, refreshTTL)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func (m *Manager) sign(uid int64, typ string, ttl time.Duration) (string, error) {
	now := time.Now()
	c := claims{
		UID: uid,
		Typ: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
}

// Parse 校验签名与有效期，返回 uid；typ 不符返回错误（refresh 不能调业务端点，反之亦然）。
func (m *Manager) Parse(tokenStr, wantTyp string) (int64, error) {
	var c claims
	token, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, errors.New("invalid token")
	}
	if c.Typ != wantTyp {
		return 0, errors.New("token type mismatch")
	}
	return c.UID, nil
}
