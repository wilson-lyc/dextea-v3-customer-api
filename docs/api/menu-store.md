# 门店菜单

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`GET /api/v1/stores/:storeId/menu`
- 模块：门店/菜单（menu）
- 鉴权：需要（见 README 鉴权说明）
- 请求格式：path
- 响应格式：统一 `{code, message, data}`

> 特例：当 `storeId` 非法（非数字或 ≤ 0）时，本接口返回 **HTTP 400 + code 40001**（非默认的 200 + code）。

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| storeId | path | int64 | 是 | 门店 ID（路径参数） |

### 请求示例
```
GET /api/v1/stores/1001/menu
```

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "menu": {
      "id": 50,
      "name": "夏季菜单",
      "description": "清爽一夏"
    },
    "groups": [
      {
        "id": 1,
        "name": "招牌",
        "sort": 1,
        "products": [
          {
            "id": 2001,
            "name": "冰美式",
            "brief": "微酸回甘",
            "price": 28.0,
            "sort": 1,
            "status": 1,
            "cover": "https://cdn.example.com/p/2001.jpg"
          }
        ]
      }
    ]
  }
}
```

#### data 字段说明
> `data` 为 `StoreMenuResponse`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data.menu | object | 菜单信息 |
| data.menu.id | int64 | 菜单 ID |
| data.menu.name | string | 菜单名称 |
| data.menu.description | string | 菜单描述 |
| data.groups | array | 商品分组列表 |
| data.groups[].id | int64 | 分组 ID |
| data.groups[].name | string | 分组名称 |
| data.groups[].sort | int | 排序 |
| data.groups[].products | array | 分组下商品列表 |
| data.groups[].products[].id | int64 | 商品 ID |
| data.groups[].products[].name | string | 商品名称 |
| data.groups[].products[].brief | string | 商品简介 |
| data.groups[].products[].price | float64 | 价格 |
| data.groups[].products[].sort | int | 排序 |
| data.groups[].products[].status | int | 状态 |
| data.groups[].products[].cover | string | 封面图 URL |

### 失败响应（特例 HTTP 400）
```json
{
  "code": 40001,
  "message": "门店ID不合法",
  "data": null
}
```

## 错误码
| 码值 | 名称 | 说明 |
| --- | --- | --- |
| 40001 | ErrBadRequest | 门店 ID 不合法（HTTP 400，特例） |
| 40400 | （内联） | 资源不存在（如门店未配置菜单） |
| 50300 | ErrMysqlDisabled | 数据库未启用 |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
