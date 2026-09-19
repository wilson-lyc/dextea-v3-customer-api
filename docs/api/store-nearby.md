# 附近门店

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`GET /api/v1/stores/nearby`
- 模块：门店（store）
- 鉴权：需要（见 README 鉴权说明）
- 请求格式：query
- 响应格式：统一 `{code, message, data}`

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| longitude | query | float64 | 是 | 经度 |
| latitude | query | float64 | 是 | 纬度 |
| distance | query | float64 | 是 | 搜索半径（需 > 0） |
| count | query | int | 是 | 返回数量（需 > 0） |

### 请求示例
```
GET /api/v1/stores/nearby?longitude=116.397&latitude=39.908&distance=5000&count=20
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
  ]
}
```

#### data 字段说明
> `data` 为门店数组，元素字段（基于 `StoreDetailItem`）如下：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | int64 | 门店 ID |
| name | string | 门店名称 |
| status | int | 营业状态 |
| province | string | 省 |
| city | string | 市 |
| district | string | 区 |
| address | string | 详细地址 |
| business_hours | string | 营业时间 |
| phone | string | 联系电话 |
| longitude | float64 | 经度 |
| latitude | float64 | 纬度 |
| distance | float64（可空） | 距当前位置距离 |
| unit | string（可空） | 距离单位：`m` 或 `km` |

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
| 40001 | ErrBadRequest | 参数绑定失败（缺参或 `distance`/`count` 不满足 `>0`） |
| 43002 | Store Service 不可用 | 门店服务下游不可用 |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
