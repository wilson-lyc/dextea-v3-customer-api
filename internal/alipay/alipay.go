// Package alipay 封装支付宝 SDK（基于 github.com/smartwalle/alipay/v3），
// 提供 OAuth 授权码换取 user_id 的能力。
//
// 本包独立于 customer 模块，可被任何需要支付宝集成的业务模块引用。
package alipay

import (
	"context"
	"errors"
	"fmt"

	"github.com/smartwalle/alipay/v3"
)

// Client 是对支付宝 SDK 的封装，当前仅暴露换取 OAuth user_id 的能力。
type Client struct {
	client *alipay.Client
}

// New 创建一个支付宝客户端。
//
// appID:           支付宝应用 AppID
// privateKey:      应用私钥（PEM 格式字符串，SDK 内部会自动解析）
// alipayPublicKey: 支付宝公钥（用于验签，可选，传空则跳过验签）
// isProduction:    是否生产环境（true=生产，false=沙箱）
func New(appID, privateKey, alipayPublicKey string, isProduction bool) (*Client, error) {
	if appID == "" {
		return nil, errors.New("alipay app_id 未配置")
	}
	if privateKey == "" {
		return nil, errors.New("alipay 应用私钥未配置")
	}

	// 使用支付宝 v3 SDK 创建客户端
	client, err := alipay.New(appID, privateKey, isProduction)
	if err != nil {
		return nil, fmt.Errorf("创建支付宝客户端失败: %w", err)
	}

	// 设置支付宝公钥用于验签（可选）
	if alipayPublicKey != "" {
		if err := client.LoadAliPayPublicKey(alipayPublicKey); err != nil {
			return nil, fmt.Errorf("加载支付宝公钥失败: %w", err)
		}
	}

	return &Client{client: client}, nil
}

// ExchangeCode 用前端授权 code 换取支付宝 user_id（即本应用的 OpenID）。
func (c *Client) ExchangeCode(ctx context.Context, code string) (string, error) {
	// 调用 alipay.system.oauth.token 接口
	p := alipay.SystemOauthToken{
		GrantType: "authorization_code",
		Code:      code,
	}

	resp, err := c.client.SystemOauthToken(ctx, p)
	if err != nil {
		return "", fmt.Errorf("调用支付宝授权接口失败: %w", err)
	}

	// 检查业务响应（SDK 的 Error 结构体嵌套在响应中）
	if resp.Code != "10000" {
		return "", fmt.Errorf("支付宝授权失败: %s(%s) %s",
			resp.SubCode, resp.Code, resp.SubMsg)
	}

	if resp.UserId == "" {
		return "", errors.New("支付宝未返回 user_id")
	}

	return resp.UserId, nil
}
