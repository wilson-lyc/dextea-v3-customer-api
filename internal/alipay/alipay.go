// Package alipay 封装支付宝 OAuth 授权码换取用户标识（OpenID）的能力。
//
// 本包不依赖任何第三方 SDK，直接使用标准库 net/http 调用支付宝 OpenAPI 网关，
// 并按支付宝 OpenAPI 规范完成 RSA2 签名与响应验签。
//
// 参考：https://opendocs.alipay.com/open-v3/ba2f3ec8_alipay.system.oauth.token
package alipay

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	productionGateway = "https://openapi.alipay.com/gateway.do"
	sandboxGateway    = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"

	oauthTokenMethod = "alipay.system.oauth.token"
	responseSuffix   = "_response"
	fieldSign        = "sign"
	fieldErrorResp   = "error_response"
	signTypeRSA2     = "RSA2"
	charsetUTF8      = "utf-8"
	formatJSON       = "JSON"
	apiVersion       = "1.0"
)

// Client 是对支付宝网关的轻量封装。
type Client struct {
	appID      string
	gateway    string
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey // 可选，用于响应验签
	httpClient *http.Client
}

// New 创建一个支付宝客户端。
//
// appID:           支付宝应用 AppID
// privateKey:      应用私钥（支持 PEM 文本或裸 base64，PKCS1/PKCS8 均自动识别）
// alipayPublicKey: 支付宝公钥（用于响应验签，可选；传空则跳过验签）
// isProduction:    true=生产网关，false=沙箱网关
func New(appID, privateKey, alipayPublicKey string, isProduction bool) (*Client, error) {
	if appID == "" {
		return nil, errors.New("alipay app_id 未配置")
	}
	if privateKey == "" {
		return nil, errors.New("alipay 应用私钥未配置")
	}

	pri, err := parsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("解析应用私钥失败: %w", err)
	}

	c := &Client{
		appID:      appID,
		privateKey: pri,
		gateway:    productionGateway,
		httpClient: http.DefaultClient,
	}
	if !isProduction {
		c.gateway = sandboxGateway
	}

	if alipayPublicKey != "" {
		pub, err := parsePublicKey(alipayPublicKey)
		if err != nil {
			return nil, fmt.Errorf("解析支付宝公钥失败: %w", err)
		}
		c.publicKey = pub
	}

	return c, nil
}

// ExchangeCode 用前端授权 code 换取支付宝用户标识（OpenID）。
//
// alipay.system.oauth.token 的 grant_type/code 是顶层表单参数（非 biz_content），
// 成功响应为 alipay_system_oauth_token_response，且不含顶层 code 字段。
func (c *Client) ExchangeCode(ctx context.Context, code string) (string, error) {
	values := url.Values{}
	values.Set("app_id", c.appID)
	values.Set("method", oauthTokenMethod)
	values.Set("format", formatJSON)
	values.Set("charset", charsetUTF8)
	values.Set("sign_type", signTypeRSA2)
	values.Set("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	values.Set("version", apiVersion)
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)

	sign, err := c.sign(values)
	if err != nil {
		return "", fmt.Errorf("生成支付宝签名失败: %w", err)
	}
	values.Set(fieldSign, sign)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.gateway, strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("构造支付宝请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用支付宝接口失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取支付宝响应失败: %w", err)
	}

	return c.parseOAuthResponse(body)
}

// parseOAuthResponse 解析网关返回的 JSON，处理 error_response 与成功响应，
// 并在配置了支付宝公钥时完成响应验签。
func (c *Client) parseOAuthResponse(body []byte) (string, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("解析支付宝响应失败: %w (body=%s)", err, string(body))
	}

	// 业务错误（error_response 中才会出现 code/sub_code 等字段）
	if errRaw, ok := raw[fieldErrorResp]; ok {
		var e oauthError
		_ = json.Unmarshal(errRaw, &e)
		return "", fmt.Errorf("支付宝授权失败: %s(%s) %s", e.SubCode, e.Code, e.SubMsg)
	}

	bizField := oauthTokenMethod + responseSuffix
	bizRaw, ok := raw[bizField]
	if !ok {
		return "", fmt.Errorf("支付宝响应缺少业务字段 %q (body=%s)", bizField, string(body))
	}

	// 响应验签（可选）：用支付宝公钥校验返回体上的 sign。
	if c.publicKey != nil {
		if signRaw, ok := raw[fieldSign]; ok && len(signRaw) > 1 {
			if err := c.verify(bizRaw, strings.Trim(string(signRaw), `"`)); err != nil {
				return "", fmt.Errorf("支付宝响应验签失败: %w", err)
			}
		}
	}

	var r oauthTokenResponse
	if err := json.Unmarshal(bizRaw, &r); err != nil {
		return "", fmt.Errorf("解析支付宝响应体失败: %w", err)
	}

	// 新版响应返回 open_id（也可能同时返回 user_id），优先 user_id，回退 open_id。
	userID := r.UserId
	if userID == "" {
		userID = r.OpenId
	}
	if userID == "" {
		return "", errors.New("支付宝未返回用户标识(open_id/user_id)")
	}
	return userID, nil
}

// sign 按支付宝规范生成 RSA2 签名：
// 将除 sign 外的所有参数按 key 升序拼成 k=v&k=v（值不 urlencode），
// 对拼接串做 SHA256WithRSA 签名后 base64 编码。
func (c *Client) sign(values url.Values) (string, error) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		for _, v := range values[k] {
			pairs = append(pairs, k+"="+v)
		}
	}
	sort.Strings(pairs)
	signStr := strings.Join(pairs, "&")

	h := sha256.New()
	h.Write([]byte(signStr))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// verify 校验支付宝响应签名：对原始业务 JSON 做 SHA256，再用支付宝公钥验签。
func (c *Client) verify(bizRaw json.RawMessage, signB64 string) error {
	sig, err := base64.StdEncoding.DecodeString(signB64)
	if err != nil {
		return fmt.Errorf("签名 base64 解码失败: %w", err)
	}
	h := sha256.New()
	h.Write(bizRaw)
	return rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, h.Sum(nil), sig)
}

// parsePrivateKey 解析应用私钥，支持 PEM 文本与裸 base64，PKCS1/PKCS8 自动识别。
func parsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	privateKey = strings.TrimSpace(privateKey)

	if strings.Contains(privateKey, "-----BEGIN") {
		block, _ := pem.Decode([]byte(privateKey))
		if block == nil {
			return nil, errors.New("无效的 PEM 格式")
		}
		return parseDERPrivateKey(block.Bytes)
	}

	// 裸 base64
	der, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("私钥 base64 解码失败: %w", err)
	}
	return parseDERPrivateKey(der)
}

func parseDERPrivateKey(der []byte) (*rsa.PrivateKey, error) {
	if k, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return k, nil
	}
	if k, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		rsaKey, ok := k.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("应用私钥不是 RSA 私钥")
		}
		return rsaKey, nil
	}
	return nil, errors.New("无法解析应用私钥（需 PKCS1 或 PKCS8 格式）")
}

// parsePublicKey 解析支付宝公钥（PKIX / SubjectPublicKeyInfo），支持 PEM 与裸 base64。
func parsePublicKey(publicKey string) (*rsa.PublicKey, error) {
	publicKey = strings.TrimSpace(publicKey)

	if strings.Contains(publicKey, "-----BEGIN") {
		block, _ := pem.Decode([]byte(publicKey))
		if block == nil {
			return nil, errors.New("无效的 PEM 格式")
		}
		return parseDERPublicKey(block.Bytes)
	}

	der, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("公钥 base64 解码失败: %w", err)
	}
	return parseDERPublicKey(der)
}

func parseDERPublicKey(der []byte) (*rsa.PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("支付宝公钥不是 RSA 公钥")
	}
	return rsaPub, nil
}

// oauthError 对应支付宝 error_response 结构。
type oauthError struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

// oauthTokenResponse 对应 alipay_system_oauth_token_response 成功结构。
type oauthTokenResponse struct {
	UserId       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	ReExpiresIn  int64  `json:"re_expires_in"`
	AuthStart    string `json:"auth_start"`
	OpenId       string `json:"open_id"`
	UnionId      string `json:"union_id"`
}
