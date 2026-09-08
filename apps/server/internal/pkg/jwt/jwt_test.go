package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestManager() *Manager {
	return NewManager("test-secret-for-unit-tests")
}

func TestIssueAndParseRoundTrip(t *testing.T) {
	m := newTestManager()
	pair, err := m.Issue(42)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if pair.ExpiresIn != int64((2 * time.Hour).Seconds()) {
		t.Errorf("ExpiresIn = %d, want 7200", pair.ExpiresIn)
	}

	uid, err := m.Parse(pair.AccessToken, TypeAccess)
	if err != nil {
		t.Fatalf("Parse access: %v", err)
	}
	if uid != 42 {
		t.Errorf("uid = %d, want 42", uid)
	}

	if _, err := m.Parse(pair.RefreshToken, TypeRefresh); err != nil {
		t.Fatalf("Parse refresh: %v", err)
	}
}

func TestParseTypeMismatch(t *testing.T) {
	m := newTestManager()
	pair, _ := m.Issue(1)

	// refresh token 调业务端点（wantTyp=access）→ 拒绝。
	if _, err := m.Parse(pair.RefreshToken, TypeAccess); err == nil {
		t.Error("refresh token 应被 access 校验拒绝")
	}
	// access token 调 refresh 端点 → 拒绝。
	if _, err := m.Parse(pair.AccessToken, TypeRefresh); err == nil {
		t.Error("access token 应被 refresh 校验拒绝")
	}
}

func TestParseExpired(t *testing.T) {
	m := newTestManager()
	// 手工签发一个已过期的 access token。
	now := time.Now().Add(-3 * time.Hour)
	c := claims{
		UID: 1,
		Typ: TypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
	if err != nil {
		t.Fatalf("sign expired: %v", err)
	}
	if _, err := m.Parse(expired, TypeAccess); err == nil {
		t.Error("过期 token 应被拒绝")
	}
}

func TestParseTampered(t *testing.T) {
	m := newTestManager()
	pair, _ := m.Issue(1)

	// 换密钥的管理器（模拟篡改签名）。
	other := NewManager("another-secret")
	if _, err := other.Parse(pair.AccessToken, TypeAccess); err == nil {
		t.Error("篡改签名的 token 应被拒绝")
	}
}

func TestParseAlgorithmConfusion(t *testing.T) {
	m := newTestManager()
	// 用 none 算法伪造（无签名）。
	c := claims{UID: 1, Typ: TypeAccess}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, c)
	forged, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}
	if _, err := m.Parse(forged, TypeAccess); err == nil {
		t.Error("none 算法伪造 token 应被拒绝")
	}
}

func TestParseGarbage(t *testing.T) {
	m := newTestManager()
	for _, s := range []string{"", "not-a-jwt", "a.b.c.d"} {
		if _, err := m.Parse(s, TypeAccess); err == nil {
			t.Errorf("垃圾串 %q 应被拒绝", s)
		}
	}
}
