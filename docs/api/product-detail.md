# 商品详情

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`GET /api/v1/products/detail`
- 模块：商品（product）
- 鉴权：需要（见 README 鉴权说明）
- 请求格式：query
- 响应格式：统一 `{code, message, data}`

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| productId | query | int64 | 是 | 商品 ID |
| storeId | query | int64 | 是 | 门店 ID（用于门店级价格/库存上下文） |

### 请求示例
```
GET /api/v1/products/detail?productId=2001&storeId=1001
```

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 2001,
    "name": "冰美式",
    "brief": "微酸回甘",
    "description": "精选咖啡豆...",
    "status": 1,
    "price": 28.0,
    "cover": { "id": 1, "url": "https://cdn.example.com/p/2001.jpg", "sort": 1, "type": 1 },
    "gallery": [
      { "id": 2, "url": "https://cdn.example.com/p/2001-2.jpg", "sort": 2, "type": 1 }
    ],
    "customizations": [
      {
        "id": 10,
        "name": "温度",
        "sort": 1,
        "status": 1,
        "options": [
          { "id": 100, "name": "冰", "price": 0, "sort": 1, "status": 1 }
        ]
      }
    ]
  }
}
```

#### data 字段说明
> `data` 为 `ProductDetailResponse`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | int64 | 商品 ID |
| name | string | 商品名称 |
| brief | string | 简介 |
| description | string | 详情描述 |
| status | int | 状态 |
| price | float64 | 价格 |
| cover | object（可空） | 封面图（`ProductImageItem`） |
| gallery | array | 图集（`ProductImageItem` 列表） |
| customizations | array | 定制项（`CustomizationItemResponse` 列表） |

**ProductImageItem**：`id`(int64)、`url`(string)、`sort`(int)、`type`(int)

**CustomizationItemResponse**：`id`(int64)、`name`(string)、`sort`(int)、`status`(int)、`options`(`CustomizationOptionItem` 列表)

**CustomizationOptionItem**：`id`(int64)、`name`(string)、`price`(float64)、`sort`(int)、`status`(int)

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
| 40001 | ErrBadRequest | 参数绑定失败（缺 `productId` / `storeId`） |
| 40400 | （内联） | 资源不存在（商品/门店不存在） |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
