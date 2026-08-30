# 逆地址编码

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`GET /api/v1/area/regeo`
- 模块：区域（area）
- 鉴权：需要（见 README 鉴权说明）
- 请求格式：query
- 响应格式：统一 `{code, message, data}`

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| longitude | query | float64 | 是 | 经度 |
| latitude | query | float64 | 是 | 纬度 |

### 请求示例
```
GET /api/v1/area/regeo?longitude=116.397&latitude=39.908
```

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "province": "北京市",
    "city": "北京市",
    "district": "朝阳区"
  }
}
```

#### data 字段说明
> `data` 为 `ReverseGeocodeResponse`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data.province | string | 省/直辖市 |
| data.city | string | 城市 |
| data.district | string | 区/县 |

### 失败响应（非 0 code）
```json
{
  "code": 43001,
  "message": "高德地图服务不可用",
  "data": null
}
```

## 错误码
| 码值 | 名称 | 说明 |
| --- | --- | --- |
| 40001 | ErrBadRequest | 参数绑定失败（缺 `longitude` / `latitude`） |
| 43001 | ErrAmapUnavailable | 高德地图服务不可用（下游） |
| 43002 | ErrAmapError | 高德地图服务返回错误（下游） |
| 43003 | ErrInvalidLocation | 经纬度参数无效（校验） |
| 50300 | ErrMysqlDisabled | 数据库未启用 |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
