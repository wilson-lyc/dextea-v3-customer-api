# 月订单列表（透传）

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-09-01 | 维护人：wilson

## 基本信息
- 路径：`GET /api/v1/orders/monthly`
- 模块：订单（order，透传代理）
- 鉴权：需要 `Authorization: Bearer {token}` 与 `X-Customer-Id` 请求头（见 README 订单类说明）
- 请求格式：query（由下游订单服务定义，透传）
- 响应格式：**下游原始报文透传**（不套用统一 `{code, message, data}`）

## 接口说明
按自然月分页查询当前顾客的历史订单，并返回当月订单总数与总金额。列表项为轻量摘要（不含订单项明细），需要明细时调用 [查询订单详情](./order-detail.md)。

## 请求结构
本接口为透传代理，query 参数原样转发至下游订单中台 `/api/v1/customer/orders/monthly`，本项目不解析。

### 请求头
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| Authorization | header | string | 是 | API 访问令牌，格式 `Bearer {token}` |
| X-Customer-Id | header | string(int64) | 是 | 顾客 ID，标识当前登录顾客 |

### 查询参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| year | query | integer | 是 | 年份，如 `2026` |
| month | query | integer | 是 | 月份，取值 1-12 |

### 请求示例
```bash
curl "http://localhost:8080/api/v1/orders/monthly?year=2026&month=4" \
  -H "Authorization: Bearer {token}" \
  -H "X-Customer-Id: 10086"
```

## 响应结构
直接透传下游订单服务的 HTTP 状态码与响应体，本文档不做二次包装。`data` 字段：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| orders | array | 当月订单列表，元素见下方「orders 数组元素」 |
| totalCount | integer | 当月订单总数 |
| totalAmount | decimal | 当月订单总金额（元），保留两位小数 |

### orders 数组元素
| 参数 | 类型 | 说明 |
| --- | --- | --- |
| id | long | 订单 ID |
| storeName | string | 门店名称 |
| createdAt | string | 下单时间，格式 `yyyy-MM-ddTHH:mm:ss` |
| totalPrice | decimal | 订单总价（元） |
| totalQuantity | integer | 订单商品总数量 |
| makingStatus | integer | 制作状态：`0` 待制作，`1` 制作中，`2` 制作完成，`3` 已取餐，`4` 已取消 |
| paymentStatus | integer | 支付状态：`0` 支付中，`1` 支付超时，`2` 已支付，`3` 退款中，`4` 已退款 |
| covers | array | 订单内商品封面图 URL 列表 |

### 响应示例
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "orders": [
      {
        "id": 1,
        "storeName": "朝阳旗舰店",
        "createdAt": "2026-04-23T15:45:00",
        "totalPrice": 25.00,
        "totalQuantity": 2,
        "makingStatus": 2,
        "paymentStatus": 2,
        "covers": ["https://example.com/a.jpg", "https://example.com/b.jpg"]
      }
    ],
    "totalCount": 12,
    "totalAmount": 320.00
  }
}
```

## 错误码
### 下游业务错误码（透传）
| 错误码 | HTTP 状态码 | 说明 |
| --- | --- | --- |
| 50000 | 500 | 系统繁忙，请稍后重试 |
| 50101 | 500 | 数据库访问异常（系统繁忙，请稍后重试） |
| 40001 | 400 | 缺少请求头 |
| 40002 | 400 | 参数缺失 / 参数校验失败（如 month 不在 1-12 之间） |
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
