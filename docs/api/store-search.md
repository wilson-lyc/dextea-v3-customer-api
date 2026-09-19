# 搜索门店

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`GET /api/v1/stores/search`
- 模块：门店（store）
- 鉴权：需要（见 README 鉴权说明）
- 请求格式：query
- 响应格式：统一 `{code, message, data}`

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| city | query | string | 否 | 城市，完全匹配 |
| keyword | query | string | 否 | 关键词，模糊匹配 |
| longitude | query | float64 | 是 | 经度（用于距离排序） |
| latitude | query | float64 | 是 | 纬度（用于距离排序） |

### 请求示例
```
GET /api/v1/stores/search?city=北京市&keyword=咖啡&longitude=116.397&latitude=39.908
```

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": [
    {
      "id": 1001,
      "name": "望京店",
      "status": 1,
      "city": "北京市",
      "district": "朝阳区",
      "address": "望京 SOHO",
      "longitude": 116.39,
      "latitude": 39.90
    }
  ]
}
```

#### data 字段说明
> `data` 为门店数组，元素字段（基于 `StoreDetailItem`）见 [store-nearby.md](./store-nearby.md#data-字段说明)。

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
| 40001 | ErrBadRequest | 参数绑定失败（缺 `longitude` / `latitude`） |
| 43002 | Store Service 不可用 | 门店服务下游不可用 |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
