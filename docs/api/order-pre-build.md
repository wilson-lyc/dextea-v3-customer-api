# 订单预构建（透传）

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`POST /api/v1/orders/pre-build`
- 模块：订单（order，透传代理）
- 鉴权：需要 `X-Customer-Id` 请求头（见 README 订单类说明）
- 请求格式：由下游订单服务定义（透传）
- 响应格式：**下游原始报文透传**（不套用统一 `{code, message, data}`）

## 请求结构
本接口为透传代理，请求体与下游订单中台 `/api/v1/customer/orders/pre-build` 契约一致，本项目不解析、不校验其字段，仅做鉴权与转发。

### 请求头
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| X-Customer-Id | header | string(int64) | 是 | 客户 ID，缺失返回 `40100` |

### 请求示例
```http
POST /api/v1/orders/pre-build
X-Customer-Id: 1001
Content-Type: application/json

{ /* 下游订单服务请求体，待补充 */ }
```

## 响应结构
直接透传下游订单服务的 HTTP 状态码与响应体，本文档不做二次包装。
> 请求/响应字段契约请参考下游订单中台文档（待补充）。

## 错误码
| 码值 | 名称 | 说明 |
| --- | --- | --- |
| 40100 | ErrUnauthorized | 缺失或非法的 `X-Customer-Id` |
| 42001 | ErrOrderServiceNotConfigured | 订单服务未配置 |
| 42002 | ErrOrderServiceUnavailable | 订单服务暂不可用（下游，可有限重试） |
| 42003 | ErrOrderServiceError | 订单服务处理失败（下游） |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
