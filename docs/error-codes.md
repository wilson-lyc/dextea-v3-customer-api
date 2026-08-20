# 错误码清单（Error Codes）

> 本文件由 `docs/异常处理机制重构方案.md` 的规范生成，汇总系统中所有通过 `bizerror` 错误中心定义的业务/系统错误码。
> 业务代码只引用下列模板，禁止散落手写文案。如需新增错误码，请在对应模块的 `errorcode.go` 中按分段规范定义。
>
> 生成时间：2026-08-20（随代码变更需同步更新）

## 错误码分段规范

| 段位 | 含义 | HTTP 映射默认 | 可重试 | 告警 |
| --- | --- | --- | --- | --- |
| `0` | 成功 | 200 | - | 否 |
| `1xxxxx` | 系统错误（基础设施/未知） | 200 + code | 否 | 是 |
| `2xxxxx` | 业务错误（领域规则不满足） | 200 + code | 否 | 否 |
| `3xxxxx` | 下游依赖错误（新分段，逐步迁移中） | 200 + code | 是 | 否 |
| `4xxxxx` | 参数 / 校验错误 | 400（鉴权 401 / 权限 403 显式指定） | 否 | 否 |
| `5xxxxx` | 限流 / 幂等 / 熔断 | 429（历史系统码 50000/50300 保留为 200） | 是 | 否 |

> 迁移期说明：历史码 `50000`（内部错误）、`50300`（数据库未启用）、`50301`（Redis 未启用）保留原值，新分段仅约束**新增**错误。

## 通用错误（internal/common/bizerror/common.go）

| 码值 | 名称 | 文案 | 类别 | 说明 |
| --- | --- | --- | --- | --- |
| `0` | ErrOK | ok | 业务 | 成功 |
| `50000` | ErrInternal | 内部错误 | 系统 | 兜底系统异常（历史码保留），HTTP 收敛不泄露细节 |
| `40000` | ErrParamInvalid | 参数错误 | 校验 | |
| `40001` | ErrBadRequest | 参数不合法 | 校验 | HTTP 400 |
| `40100` | ErrUnauthorized | 未登录或登录已过期 | 校验 | HTTP 401（显式） |
| `40301` | ErrForbidden | 无权限访问 | 校验 | HTTP 403（显式） |
| `50001` | ErrTooManyRequests | 请求过于频繁，请稍后重试 | 限流 | HTTP 429 |
| `50002` | ErrCircuitBreaker | 服务繁忙，请稍后重试 | 限流 | 下游熔断触发 |
| `50003` | ErrDuplicateSubmit | 请勿重复提交 | 限流 | 幂等冲突 |

## 基础设施错误（internal/common/bizerror/errorcode.go）

| 码值 | 名称 | 文案 | 类别 |
| --- | --- | --- | --- |
| `50300` | ErrMysqlDisabled | 数据库未启用 | 系统 |
| `50301` | ErrRedisDisabled | Redis 未启用 | 系统 |

## 订单模块（internal/biz/order/errorcode.go）

| 码值 | 名称 | 文案 | 类别 | 说明 |
| --- | --- | --- | --- | --- |
| `42001` | ErrOrderServiceNotConfigured | 订单服务未配置 | 下游 | Nacos/静态地址皆不可用 |
| `42002` | ErrOrderServiceUnavailable | 订单服务暂不可用 | 下游 | 网络/超时瞬时失败，支持有限重试 |
| `42003` | ErrOrderServiceError | 订单服务处理失败 | 下游 | 下游返回非预期响应 |

## 客服/顾客模块（internal/biz/customer/errorcode.go）

| 码值 | 名称 | 文案 | 类别 |
| --- | --- | --- | --- |
| `41001` | ErrPlatformInvalid | 不支持的登录平台 | 校验 |
| `41002` | ErrPlatformNotSupported | 该登录平台暂未开放 | 校验 |
| `41003` | ErrAlipayNotConfigured | 支付宝登录未配置 | 下游 |
| `41004` | ErrAlipayAuthFailed | 支付宝授权失败 | 下游 |
| `44001` | ErrCustomerDataEmpty | 客服数据为空 | 下游 |

## 区域/门店模块（internal/biz/area/errorcode.go）

| 码值 | 名称 | 文案 | 类别 |
| --- | --- | --- | --- |
| `43001` | ErrAmapUnavailable | 高德地图服务不可用 | 下游 |
| `43002` | ErrAmapError | 高德地图服务返回错误 | 下游 |
| `43003` | ErrInvalidLocation | 经纬度参数无效 | 校验 |

## 其他模块内联错误

| 码值 | 文案（模板） | 使用位置 | 类别 |
| --- | --- | --- | --- |
| `40400` | 资源不存在（商品/菜单等具体文案由调用处覆盖） | product/service.go、menu/service.go | 业务 |

---

## 使用约定

- 业务代码统一引用上述模板，例如：`bizerror.ErrBadRequest`、`bizerror.New(bizerror.ErrInternal)`、`bizerror.NewWith(bizerror.ErrOrderServiceUnavailable, bizerror.WithCause(err))`。
- 携带根因用 `bizerror.WithCause(err)`，携带排障上下文用 `bizerror.WithField("orderId", id)`，二者仅出现在日志，不回显客户端。
- 全局 `ExceptionInterceptor` 与 `response.ErrorOf` 会根据 `Kind` 自动推导 HTTP 状态码、是否告警，并在系统类错误时记录 error 级日志 + 根因。
- 跨服务（订单中台）透传保持 `200 + code` 契约，前端按 `code` 判断；校验类（4xxxxx）暴露真实 4xx 便于边缘识别。
