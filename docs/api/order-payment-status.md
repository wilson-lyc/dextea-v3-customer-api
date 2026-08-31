# 订单支付状态（透传）

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-09-01 | 维护人：wilson

## 基本信息
- 路径：`GET /api/v1/orders/:orderId/payment-status`
- 模块：订单（order，透传代理）
- 鉴权：需要 `Authorization: Bearer {token}` 与 `X-Customer-Id` 请求头（见 README 订单类说明）
- 请求格式：path（由下游订单服务定义，透传）
- 响应格式：**下游原始报文透传**（不套用统一 `{code, message, data}`）

## 接口说明
轻量接口，仅返回订单的支付状态，供客户端在拉起支付后轮询。仅能查询归属当前顾客（`X-Customer-Id`）的订单。特殊行为：当本地支付状态为「支付中」时，服务端会主动向支付渠道查询一次真实状态并回写本地，因此该接口的返回可能领先于本地库中的旧状态。需要订单其他信息时调用 [查询订单详情](./order-detail.md)。

## 请求结构
本接口为透传代理，`orderId` 路径参数原样转发至下游订单中台 `/api/v1/customer/orders/:orderId/payment-status`，本项目不解析。

### 路径参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| orderId | path | long | 是 | 订单 ID |

### 请求头
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| Authorization | header | string | 是 | API 访问令牌，格式 `Bearer {token}` |
| X-Customer-Id | header | string(int64) | 是 | 顾客 ID，标识当前登录顾客 |

### 请求示例
```bash
curl http://localhost:8080/api/v1/orders/1/payment-status \
  -H "Authorization: Bearer {token}" \
  -H "X-Customer-Id: 10086"
```

## 响应结构
直接透传下游订单服务的 HTTP 状态码与响应体，本文档不做二次包装。`data` 字段：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| paymentStatus | integer | 支付状态：`0` 支付中，`1` 支付超时，`2` 已支付，`3` 退款中，`4` 已退款 |

### 响应示例
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "paymentStatus": 2
  }
}
```

## 错误码
### 下游业务错误码（透传）
| 错误码 | HTTP 状态码 | 说明 |
| --- | --- | --- |
| 50000 | 500 | 系统繁忙，请稍后重试 |
| 50101 | 500 | 数据库访问异常（系统繁忙，请稍后重试） |
| 21016 | 400 | 订单不存在 |
| 21017 | 400 | 该订单不属于当前顾客 |
| 22004 | 400 | 支付宝交易查询失败 |
| 40001 | 400 | 缺少请求头 |
| 40002 | 400 | 参数缺失 / 参数类型不正确 |
| 40100 | 401 | 未登录 |

### 网关自身错误码
| 码值 | 名称 | 说明 |
| --- | --- | --- |
| 40100 | ErrUnauthorized | 缺失或非法的 `X-Customer-Id` |
| 42001 | ErrOrderServiceNotConfigured | 订单服务未配置 |
| 42002 | ErrOrderServiceUnavailable | 订单服务暂不可用（下游，可有限重试） |
| 42003 | ErrOrderServiceError | 订单服务处理失败（下游） |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
