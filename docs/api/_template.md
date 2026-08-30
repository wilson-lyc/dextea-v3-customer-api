# [接口名称]

> 维护说明：本文件由人工维护。字段变更请同步更新。
> 最后更新：YYYY-MM-DD | 维护人：CodeBuddy

## 基本信息
- 路径：`METHOD /api/v1/...`
- 模块：xxx
- 鉴权：需要 / 不需要（见 README 鉴权说明）
- 请求格式：query / json / path
- 响应格式：统一 `{code, message, data}`（订单类为下游透传，见 README）

## 请求结构
### 请求参数
| 参数 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| field | query / json / path | string | 是 | 参数说明 |

### 请求示例
```json
{
  "field": "value"
}
```
> query 参数示例：`GET /api/v1/...?field=value`

## 响应结构
### 成功响应（HTTP 200）
```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

#### data 字段说明
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data.field | string | 字段说明 |

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
| 40001 | ErrBadRequest | 参数绑定失败 |

> 完整错误码定义见 [../error-codes.md](../error-codes.md)。
