# 城市列表

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：2026-08-30 | 维护人：CodeBuddy

## 基本信息
- 路径：`GET /api/v1/area/cities`
- 模块：区域（area）
- 鉴权：需要（见 README 鉴权说明）
- 请求格式：无
- 响应格式：统一 `{code, message, data}`

## 请求结构
### 请求参数
无。

### 请求示例
```
GET /api/v1/area/cities
```

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": [
    { "letter": "B", "cities": ["北京市", "北海市"] },
    { "letter": "S", "cities": ["上海市", "深圳市"] }
  ]
}
```

#### data 字段说明
> `data` 为 `CityLetterGroup` 数组（按首字母分组）：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data[].letter | string | 首字母分组键 |
| data[].cities | array<string> | 该分组下的城市名称列表 |

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
| 43002 | Store Service 不可用 | 城市列表下游服务不可用 |
| 50000 | ErrInternal | 系统异常兜底 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
