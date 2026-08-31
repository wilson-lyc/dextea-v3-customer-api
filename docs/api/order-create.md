# 创建订单（透传）

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-09-01 | 维护人：wilson

## 基本信息
- 路径：`POST /api/v1/orders`
- 模块：订单（order，透传代理）
- 鉴权：需要 `Authorization: Bearer {token}` 与 `X-Customer-Id` 请求头（见 README 订单类说明）
- 请求格式：由下游订单服务定义（透传）
- 响应格式：**下游原始报文透传**（不套用统一 `{code, message, data}`）

## 接口说明
顾客提交订单并创建支付交易。通过 `idempotencyKey` 保证幂等：同一 key 重复提交返回 `21014`，不会重复建单。`paymentMethod = 1`（支付宝）时会调用支付宝下单并返回 `tradeNo`，客户端凭 `tradeNo` 拉起支付；请在 `paymentExpiredAt` 前完成支付，超时后订单会被标记为支付超时。订单创建后处于「支付中」状态，[订单预构建](./order-pre-build.md) 的校验逻辑在本接口内重复执行，不可售项会放入 `unavailable`，可售项才会进入订单。

## 请求结构
本接口为透传代理，请求体与下游订单中台 `/api/v1/customer/orders` 契约一致，本项目不解析、不校验其字段，仅做鉴权与转发。

### 请求头
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| Authorization | header | string | 是 | API 访问令牌，格式 `Bearer {token}` |
| X-Customer-Id | header | string(int64) | 是 | 顾客 ID，标识当前登录顾客 |

### Body 参数（application/json）
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| storeId | long | 是 | 门店 ID |
| diningMethod | integer | 是 | 用餐方式：`1` - 堂食，`2` - 外带 |
| source | integer | 是 | 订单来源：`0` - 线下，`1` - 支付宝，`2` - 微信，`3` - APP |
| paymentMethod | integer | 是 | 支付方式：`0` - 现金，`1` - 支付宝，`2` - 微信；支付宝/微信支付要求顾客已绑定对应 openId |
| idempotencyKey | string | 是 | 幂等键，客户端生成（如 UUID 去中划线），同一订单的重试请求必须复用同一值，最长 64 字符 |
| note | string | 否 | 订单备注，最长 500 字 |
| items | array | 是 | 订单明细列表，不能为空，见下方「items 数组元素」 |

### items 数组元素
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| skuId | string | 是 | 商品 SKU 编码，如 `1#1_1-2_6-3_7` |
| quantity | integer | 是 | 购买数量，必须大于 0 |

### 请求示例
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer {token}" \
  -H "X-Customer-Id: 10086" \
  -H "Content-Type: application/json" \
  -d '{
    "storeId": 1,
    "diningMethod": 2,
    "source": 1,
    "paymentMethod": 1,
    "idempotencyKey": "f47ac10b58cc4372a5670e02b2c3d479",
    "note": "少放辣椒，不要香菜",
    "items": [
      { "skuId": "1#1_1-2_6-3_7", "quantity": 1 }
    ]
  }'
```

## 响应结构
直接透传下游订单服务的 HTTP 状态码与响应体，本文档不做二次包装。`data` 字段：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| id | long | 订单 ID |
| orderNo | string | 订单号 |
| tradeNo | string | 支付渠道交易号，用于客户端拉起支付；现金等方式可能为空 |
| paymentExpiredAt | string | 支付过期时间，格式 `yyyy-MM-ddTHH:mm:ss` |
| available | array | 可售明细列表，元素结构同 [订单预构建](./order-pre-build.md) 响应中的 items |
| unavailable | array | 不可售明细列表（未进入订单），元素结构同 available |
| totalQuantity | integer | 订单商品总数量 |
| totalPrice | decimal | 订单总价（元），保留两位小数 |

### 响应示例
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": 1,
    "orderNo": "20150320010101001",
    "tradeNo": "2015042321001004720200028594",
    "paymentExpiredAt": "2026-04-23T15:45:00",
    "available": [
      {
        "skuId": "1#1_1-2_6-3_7",
        "quantity": 1,
        "product": "招牌奶茶",
        "customization": "少冰 / 少甜 / 茉莉花茶",
        "cover": "https://example.com/a.jpg",
        "unitPrice": 12.50,
        "totalPrice": 12.50
      }
    ],
    "unavailable": [],
    "totalQuantity": 1,
    "totalPrice": 12.50
  }
}
```

## 错误码
### 下游业务错误码（透传）
| 错误码 | HTTP 状态码 | 说明 |
| --- | --- | --- |
| 50000 | 500 | 系统繁忙，请稍后重试 |
| 50101 | 500 | 数据库访问异常（系统繁忙，请稍后重试） |
| 21001 | 400 | 顾客不存在 |
| 21002 | 400 | 门店不存在 |
| 21003 | 400 | 商品不存在 |
| 21006 | 400 | 门店未营业 |
| 21007 | 400 | 顾客不可用 |
| 21008 | 400 | 非法的 SKU |
| 21010 | 400 | 非法的用餐方式 |
| 21011 | 400 | 非法的订单来源 |
| 21012 | 400 | 非法的支付方式（含未绑定对应渠道 openId） |
| 21013 | 400 | 订单项数量非法 |
| 21014 | 400 | 重复提交，请勿重复创建订单（idempotencyKey 冲突） |
| 22001 | 400 | 支付宝创建交易失败 |
| 22002 | 400 | 支付宝配置缺失 |
| 40001 | 400 | 缺少请求头 |
| 40002 | 400 | 参数缺失 / 参数校验失败 / 请求体格式错误 |
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
