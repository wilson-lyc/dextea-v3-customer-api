// Package jwt 提供轻量的 JWT（HS256）生成能力，仅依赖标准库，
// 避免为单一登录场景引入额外的第三方依赖。
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// 解析阶段可能出现的错误。
var (
	// ErrTokenMalformed 表示 token 格式不合法（不是由 header.payload.signature 三段组成）。
	ErrTokenMalformed = errors.New("token 格式不合法")
	// ErrTokenSignature 表示签名校验失败，token 可能被篡改。
	ErrTokenSignature = errors.New("token 签名校验失败")
	// ErrTokenExpired 表示 token 已过期。
	ErrTokenExpired = errors.New("token 已过期")
)

// Claims 是写入 JWT 载荷的声明。
type Claims struct {
	UID      int64  `json:"uid"`
	Platform string `json:"platform"`

	// 以下字段仅用于解析后的校验，不参与生成。
	IssuedAt int64 `json:"iat,omitempty"`
	ExpiresAt int64 `json:"exp,omitempty"`
}

// Generate 使用 HMAC-SHA256 生成一个 HS256 的 JWT 字符串。
// expireSeconds 为令牌有效期（秒）。
func Generate(secret string, claims Claims, expireSeconds int) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	now := time.Now()
	payload := map[string]interface{}{
		"uid":      claims.UID,
		"platform": claims.Platform,
		"iat":      now.Unix(),
		"exp":      now.Add(time.Duration(expireSeconds) * time.Second).Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signingInput := header + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

// Parse 使用 HMAC-SHA256 校验 token 的签名并校验过期时间，成功返回载荷中的 Claims。
//
// token 必须为形如 header.payload.signature 的三段式 JWT；签名不匹配、格式非法
// 或已过期时分别返回 ErrTokenSignature、ErrTokenMalformed、ErrTokenExpired。
func Parse(secret string, token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrTokenMalformed
	}

	// 验签：用同一 secret 对 header.payload 重新计算签名并与传入签名比对。
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, ErrTokenSignature
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrTokenMalformed
	}
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrTokenMalformed
	}

	if claims.ExpiresAt != 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}
