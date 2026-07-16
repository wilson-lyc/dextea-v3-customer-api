package customer

import (
	"bytes"
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

const alipayOAuthMethod = "alipay.system.oauth.token"
const defaultAlipayGateway = "https://openapi.alipay.com/gateway.do"

// alipayClient 封装 alipay.system.oauth.token 的调用，用于把前端授权 code 换取支付宝 user_id（OpenID）。
type alipayClient struct {
	appID      string
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	gateway    string
	httpClient *http.Client
}

func newAlipayClient(appID, privateKeyPEM, publicKeyPEM, gateway string) (*alipayClient, error) {
	if appID == "" {
		return nil, errors.New("alipay app_id 未配置")
	}
	priv, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析支付宝应用私钥失败: %w", err)
	}
	c := &alipayClient{
		appID:      appID,
		privateKey: priv,
		gateway:    gateway,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	if c.gateway == "" {
		c.gateway = defaultAlipayGateway
	}
	if publicKeyPEM != "" {
		pub, err := parsePublicKey(publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("解析支付宝公钥失败: %w", err)
		}
		c.publicKey = pub
	}
	return c, nil
}

// exchangeCode 用授权 code 调用支付宝换取 user_id（即本应用的 OpenID）。
func (c *alipayClient) exchangeCode(ctx context.Context, code string) (string, error) {
	params := map[string]string{
		"app_id":     c.appID,
		"method":     alipayOAuthMethod,
		"format":     "JSON",
		"charset":    "utf-8",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
		"grant_type": "authorization_code",
		"code":       code,
	}
	sign, err := rsa2Sign(params, c.privateKey)
	if err != nil {
		return "", fmt.Errorf("支付宝请求签名失败: %w", err)
	}
	params["sign"] = sign

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.gateway, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求支付宝网关失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取支付宝响应失败: %w", err)
	}

	return parseOAuthResponse(body, c.publicKey)
}

type alipayOAuthResult struct {
	Response      json.RawMessage `json:"alipay_system_oauth_token_response"`
	Sign          string          `json:"sign"`
	ErrorResponse struct {
		Code    string `json:"code"`
		Msg     string `json:"msg"`
		SubCode string `json:"sub_code"`
		SubMsg  string `json:"sub_msg"`
	} `json:"error_response"`
}

func parseOAuthResponse(body []byte, publicKey *rsa.PublicKey) (string, error) {
	var result alipayOAuthResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析支付宝响应失败: %w", err)
	}
	if result.ErrorResponse.Code != "" {
		return "", fmt.Errorf("支付宝返回错误: %s(%s) %s", result.ErrorResponse.SubCode, result.ErrorResponse.Code, result.ErrorResponse.SubMsg)
	}
	if publicKey != nil && result.Sign != "" {
		ok, verr := rsa2Verify(strings.TrimSpace(string(result.Response)), result.Sign, publicKey)
		if verr != nil {
			return "", fmt.Errorf("校验支付宝响应签名失败: %w", verr)
		}
		if !ok {
			return "", errors.New("支付宝响应签名校验不通过")
		}
	}

	var tokenResp struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(result.Response, &tokenResp); err != nil {
		return "", fmt.Errorf("解析支付宝授权信息失败: %w", err)
	}
	if tokenResp.UserID == "" {
		return "", errors.New("支付宝未返回 user_id")
	}
	return tokenResp.UserID, nil
}

func rsa2Sign(params map[string]string, key *rsa.PrivateKey) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b bytes.Buffer
	for _, k := range keys {
		if params[k] == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("&")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(params[k])
	}

	h := sha256.New()
	h.Write(b.Bytes())
	digest := h.Sum(nil)

	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func rsa2Verify(content, signBase64 string, key *rsa.PublicKey) (bool, error) {
	sig, err := base64.StdEncoding.DecodeString(signBase64)
	if err != nil {
		return false, err
	}
	h := sha256.New()
	h.Write([]byte(content))
	digest := h.Sum(nil)
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest, sig); err != nil {
		return false, nil
	}
	return true, nil
}

func parsePrivateKey(raw string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(normalizeKey(raw, "RSA PRIVATE KEY"))
	if block == nil {
		return nil, errors.New("私钥格式无效")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	keyIfc, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := keyIfc.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("私钥不是 RSA 私钥")
	}
	return key, nil
}

func parsePublicKey(raw string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(normalizeKey(raw, "PUBLIC KEY"))
	if block == nil {
		return nil, errors.New("公钥格式无效")
	}
	keyIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := keyIfc.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("公钥不是 RSA 公钥")
	}
	return key, nil
}

// normalizeKey 把 .env 中的密钥（可以是裸 base64，也可能已经是 PEM）统一规整为 PEM 字节。
func normalizeKey(raw, pemType string) []byte {
	if strings.Contains(raw, "-----BEGIN") {
		return []byte(raw)
	}
	clean := strings.Join(strings.Fields(raw), "")
	var sb strings.Builder
	sb.WriteString("-----BEGIN ")
	sb.WriteString(pemType)
	sb.WriteString("-----\n")
	for i := 0; i < len(clean); i += 64 {
		end := i + 64
		if end > len(clean) {
			end = len(clean)
		}
		sb.WriteString(clean[i:end])
		sb.WriteString("\n")
	}
	sb.WriteString("-----END ")
	sb.WriteString(pemType)
	sb.WriteString("-----\n")
	return []byte(sb.String())
}
