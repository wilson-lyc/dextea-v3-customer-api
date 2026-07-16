// Package jwt 提供轻量的 JWT（HS256）生成能力，仅依赖标准库，
// 避免为单一登录场景引入额外的第三方依赖。
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"
)

// Claims 是写入 JWT 载荷的声明。
type Claims struct {
	UID      int64  `json:"uid"`
	Platform string `json:"platform"`
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
