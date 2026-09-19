# 门店详情

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`GET /api/v1/stores/detail`
- 模块：门店（store）
- 鉴权：需要（见 README 鉴权说明）
- 请求格式：query
- 响应格式：统一 `{code, message, data}`

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | query | int64 | 是 | 门店 ID |
| longitude | query | float64 | 是 | 经度（用于计算距离） |
| latitude | query | float64 | 是 | 纬度（用于计算距离） |

### 请求示例
```
GET /api/v1/stores/detail?id=1001&longitude=116.397&latitude=39.908
```

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1001,
    "name": "望京店",
    "status": 1,
    "province": "北京市",
    "city": "北京市",
    "district": "朝阳区",
    "address": "望京 SOHO",
    "business_hours": "09:00-22:00",
    "phone": "010-12345678",
    "longitude": 116.39,
    "latitude": 39.90,
    "distance": 320.5,
    "unit": "m"
  }
}
```

#### data 字段说明
> `data` 为单个门店对象（基于 `StoreDetailItem`），字段见 [store-nearby.md](./store-nearby.md#data-字段说明)。

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
| 40001 | ErrBadRequest | 参数绑定失败（缺 `id` / `longitude` / `latitude`） |
| 43002 | Store Service 不可用 | 门店服务下游不可用 |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
