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
	"strconv"
	"strings"
	"time"
)

const (
	productionGateway = "https://openapi.alipay.com"
	sandboxGateway    = "https://openapi-sandbox.dl.alipaydev.com"
	v3OAuthTokenPath = "/v3/alipay/system/oauth/token"
	authScheme = "ALIPAY-SHA256withRSA"
)

type Client struct {
	appID      string
	gateway    string
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	httpClient *http.Client
}

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

func (c *Client) ExchangeCode(ctx context.Context, code string) (string, error) {
	bodyBytes, err := json.Marshal(map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	})
	if err != nil {
		return "", fmt.Errorf("构造支付宝请求体失败: %w", err)
	}
	body := string(bodyBytes)

	auth, err := c.buildAuthorization(http.MethodPost, v3OAuthTokenPath, body)
	if err != nil {
		return "", fmt.Errorf("生成支付宝签名失败: %w", err)
	}

	reqURL := c.gateway + v3OAuthTokenPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("构造支付宝请求失败: %w", err)
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用支付宝接口失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取支付宝响应失败: %w", err)
	}

	// 配置了支付宝公钥时，对响应做验签（网关/非业务错误可能无签名头，此时跳过）。
	if err := c.verifyResponse(respBody, resp.Header); err != nil {
		return "", fmt.Errorf("支付宝响应验签失败: %w (body=%s)", err, string(respBody))
	}

	return c.parseOAuthResponseV3(respBody)
}

// buildAuthorization 按 v3 规范生成 Authorization 头值：
//
//	authString = "app_id=...,nonce=...,timestamp=..."
//	待签名串    = authString\n<METHOD>\n<URI>\n<body>\n
//	Authorization = "ALIPAY-SHA256withRSA <authString>,sign=<RSA2签名>"
func (c *Client) buildAuthorization(method, uri, body string) (string, error) {
	nonce := newNonce()
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	authString := "app_id=" + c.appID + ",nonce=" + nonce + ",timestamp=" + timestamp
	content := authString + "\n" + method + "\n" + uri + "\n" + body + "\n"

	sig, err := c.sign(content)
	if err != nil {
		return "", err
	}
	return authScheme + " " + authString + ",sign=" + sig, nil
}

// sign 对内容做 SHA256WithRSA 签名后 base64 编码。
func (c *Client) sign(content string) (string, error) {
	h := sha256.New()
	h.Write([]byte(content))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// verifyResponse 用响应头中的 alipay-signature / alipay-timestamp / alipay-nonce 验签响应体。
// 待验签串 = timestamp\nnonce\nbody\n
func (c *Client) verifyResponse(body []byte, headers http.Header) error {
	if c.publicKey == nil {
		return nil
	}
	sig := headers.Get("alipay-signature")
	ts := headers.Get("alipay-timestamp")
	nonce := headers.Get("alipay-nonce")
	if sig == "" || ts == "" || nonce == "" {
		return nil
	}
	content := ts + "\n" + nonce + "\n" + string(body) + "\n"
	return c.verify(content, sig)
}

// verify 用支付宝公钥校验 SHA256WithRSA 签名。
func (c *Client) verify(content, signB64 string) error {
	sig, err := base64.StdEncoding.DecodeString(signB64)
	if err != nil {
		return fmt.Errorf("签名 base64 解码失败: %w", err)
	}
	h := sha256.New()
	h.Write([]byte(content))
	return rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, h.Sum(nil), sig)
}

// parseOAuthResponseV3 解析 v3 的 JSON 响应。
// 成功响应为 camelCase 字段（openId / userId 等），无 code 字段；
// 错误响应含 code / subCode 字段。
func (c *Client) parseOAuthResponseV3(body []byte) (string, error) {
	var errResp oauthErrorV3
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.SubCode != "" || (errResp.Code != "" && errResp.Code != "10000") {
			return "", fmt.Errorf("支付宝授权失败: %s(%s) %s", errResp.SubCode, errResp.Code, errResp.SubMsg)
		}
	}

	var r oauthTokenResponseV3
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("解析支付宝响应失败: %w (body=%s)", err, string(body))
	}

	// 优先 user_id，回退 open_id。
	userID := r.UserId
	if userID == "" {
		userID = r.OpenId
	}
	if userID == "" {
		return "", errors.New("支付宝未返回用户标识(openId/userId)")
	}
	return userID, nil
}

// newNonce 生成 RFC4122 v4 UUID 作为随机串。
func newNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
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

// oauthErrorV3 对应 v3 错误响应（snake_case）。
type oauthErrorV3 struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

// oauthTokenResponseV3 对应 v3 成功响应（snake_case）。
type oauthTokenResponseV3 struct {
	UserId       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	ReExpiresIn  int64  `json:"re_expires_in"`
	AuthStart    string `json:"auth_start"`
	OpenId       string `json:"open_id"`
	UnionId      string `json:"union_id"`
}
