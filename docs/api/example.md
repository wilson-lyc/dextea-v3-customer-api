# 统一格式示例：客户登录

> 本文件是接口文档的**完整示例**，演示统一格式应如何填写。
> 实际接口请复制 [`_template.md`](./_template.md) 后按此风格编写。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`POST /api/v1/customers/login`
- 模块：客户（customer）
- 鉴权：不需要（在白名单内，见 README 鉴权说明）
- 请求格式：json（body）
- 响应格式：统一 `{code, message, data}`

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| code | json | string | 是 | 第三方平台授权码（如支付宝 `auth_code` / 微信 `code`） |
| platform | json | string | 是 | 登录平台枚举：`weixin` 或 `alipay` |

### 请求示例
```json
{
  "code": "auth_code_xxxxxxxx",
  "platform": "alipay"
}
```

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "customer": {
      "id": 1001,
      "name": "张三",
      "email": "",
      "phone": "138****0000",
      "status": 1,
      "created_at": "2026-01-01T08:00:00Z",
      "updated_at": "2026-01-01T08:00:00Z"
    }
  }
}
```

#### data 字段说明
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data.token | string | 登录凭证 JWT，后续请求携带 |
| data.customer | object | 客户账号信息 |
| data.customer.id | int64 | 客户 ID |
| data.customer.name | string | 昵称 |
| data.customer.email | string | 邮箱（可能为空） |
| data.customer.phone | string | 手机号 |
| data.customer.status | int | 账号状态（1=正常） |
| data.customer.created_at | string | 创建时间（RFC3339） |
| data.customer.updated_at | string | 更新时间（RFC3339） |

> 注：出于安全，`WeixinOpenID` / `AlipayOpenID` / `Password` 字段不返回（序列化忽略）。

### 失败响应（非 0 code）
```json
{
  "code": 41001,
  "message": "不支持的登录平台",
  "data": null
}
```

## 错误码
| 码值 | 名称 | 说明 |
| --- | --- | --- |
| 40001 | ErrBadRequest | 参数绑定失败（缺 `code` / `platform`，或非 JSON 请求体） |
| 41001 | ErrPlatformInvalid | `platform` 不在 `{weixin, alipay}` 枚举内 |
| 41003 | ErrAlipayNotConfigured | 支付宝登录未配置 |
| 41004 | ErrAlipayAuthFailed | 支付宝授权失败（换取 OpenID 失败） |
| 50000 | ErrInternal | 系统异常兜底（文案收敛，根因见日志） |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
