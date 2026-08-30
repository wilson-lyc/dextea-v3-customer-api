# 商品状态

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`POST /api/v1/products/status`
- 模块：商品（product）
- 鉴权：需要（见 README 鉴权说明）
- 请求格式：json（body）
- 响应格式：统一 `{code, message, data}`

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| storeId | json | int64 | 是 | 门店 ID |
| productId | json | []int64 | 是 | 商品 ID 列表 |

### 请求示例
```json
{
  "storeId": 1001,
  "productId": [2001, 2002, 2003]
}
```

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": [
    { "productId": 2001, "name": "冰美式", "status": 1 },
    { "productId": 2002, "name": "拿铁", "status": 0 }
  ]
}
```

#### data 字段说明
> `data` 为 `ProductStoreStatusItem` 数组：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data[].productId | int64 | 商品 ID |
| data[].name | string | 商品名称 |
| data[].status | int | 该门店下商品状态（如 0=下架 / 1=在售） |

### 失败响应（非 0 code）
```json
{
  "code": 40001,
  "message": "参数不合法",
  "data": null
}
```

## 错误码
| 码值 | 名称 | 说明 |
| --- | --- | --- |
| 40001 | ErrBadRequest | 参数绑定失败（缺 `storeId` / `productId`） |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
