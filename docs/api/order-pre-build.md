# 订单预构建（透传）

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-09-01 | 维护人：wilson

## 基本信息
- 路径：`POST /api/v1/orders/pre-build`
- 模块：订单（order，透传代理）
- 鉴权：需要 `Authorization: Bearer {token}` 与 `X-Customer-Id` 请求头（见 README 订单类说明）
- 请求格式：由下游订单服务定义（透传）
- 响应格式：**下游原始报文透传**（不套用统一 `{code, message, data}`）

## 接口说明
顾客下单前预构建订单：校验顾客、门店与商品可售状态，并试算价格。服务端会逐项校验 `items` 中的 SKU，可售项放入 `available` 并回填商品信息与价格，不可售项放入 `unavailable`。预构建不产生订单，不落库；正式下单需调用 [创建订单](./order-create.md)。

## 请求结构
本接口为透传代理，请求体与下游订单中台 `/api/v1/customer/orders/pre-build` 契约一致，本项目不解析、不校验其字段，仅做鉴权与转发。

### 请求头
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| Authorization | header | string | 是 | API 访问令牌，格式 `Bearer {token}` |
| X-Customer-Id | header | string(int64) | 是 | 顾客 ID，标识当前登录顾客 |

### Body 参数（application/json）
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| storeId | long | 是 | 门店 ID |
| items | array | 是 | 订单明细列表，不能为空，见下方「items 数组元素」 |

### items 数组元素
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| skuId | string | 是 | 商品 SKU 编码，如 `1#1_1-2_6-3_7` |
| quantity | integer | 是 | 购买数量，必须大于 0 |

### 请求示例
```bash
curl -X POST http://localhost:8080/api/v1/orders/pre-build \
  -H "Authorization: Bearer {token}" \
  -H "X-Customer-Id: 10086" \
  -H "Content-Type: application/json" \
  -d '{
    "storeId": 1,
    "items": [
      { "skuId": "1#1_1-2_6-3_7", "quantity": 1 },
      { "skuId": "2#1_3", "quantity": 2 }
    ]
  }'
```

## 响应结构
直接透传下游订单服务的 HTTP 状态码与响应体。`data` 字段：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| available | array | 可售明细列表，元素结构见下表 |
| unavailable | array | 不可售明细列表（售罄、下架、非法 SKU 等），元素结构同 available |
| totalQuantity | integer | 可售项商品总数量 |
| totalPrice | decimal | 可售项总价（元），保留两位小数 |

`available` / `unavailable` 数组元素：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| skuId | string | 商品 SKU 编码 |
| quantity | integer | 购买数量 |
| product | string | 商品名称（响应填充，请求无需传） |
| customization | string | 商品客制化描述（响应填充，请求无需传） |
| cover | string | 商品封面图 URL（响应填充，请求无需传） |
| unitPrice | decimal | 单价（元）（响应填充，请求无需传） |
| totalPrice | decimal | 小计（元）（响应填充，请求无需传） |

### 响应示例
```json
{
  "code": 0,
  "message": "成功",
  "data": {
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
    "unavailable": [
      {
        "skuId": "2#1_3",
        "quantity": 2,
        "product": "杨枝甘露",
        "customization": "标准",
        "cover": "https://example.com/b.jpg",
        "unitPrice": 18.00,
        "totalPrice": 36.00
      }
    ],
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
| 21013 | 400 | 订单项数量非法 |
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
