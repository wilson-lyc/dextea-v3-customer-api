# 接口文档（API Reference）

> 本目录为 **dextea-customer-api** 的接口文档，**由人工维护（Markdown）**。
> 每个接口对应一个 `.md` 文件，统一格式见 [`_template.md`](./_template.md)，完整示例见 [`example.md`](./example.md)。
> 字段发生变动时请同步更新对应接口文件，并在文件头「最后更新」处标注日期与维护人。

## 目录索引

| 模块 | 接口 | 文档 |
| --- | --- | --- |
| 客户 | POST /api/v1/customers/login（客户登录） | [customer-login.md](./customer-login.md) |
| 门店 | GET /api/v1/stores/nearby（附近门店） | [store-nearby.md](./store-nearby.md) |
| 门店 | GET /api/v1/stores/search（搜索门店） | [store-search.md](./store-search.md) |
| 门店 | GET /api/v1/stores/detail（门店详情） | [store-detail.md](./store-detail.md) |
| 门店 | GET /api/v1/stores/:storeId/menu（门店菜单） | [menu-store.md](./menu-store.md) |
| 商品 | GET /api/v1/products/detail（商品详情） | [product-detail.md](./product-detail.md) |
| 商品 | POST /api/v1/products/status（商品状态） | [product-status.md](./product-status.md) |
| 区域 | GET /api/v1/area/regeo（逆地址编码） | [area-regeo.md](./area-regeo.md) |
| 区域 | GET /api/v1/area/cities（城市列表） | [area-cities.md](./area-cities.md) |
| 订单（透传） | POST /api/v1/orders/pre-build（订单预构建） | [order-pre-build.md](./order-pre-build.md) |
| 订单（透传） | POST /api/v1/orders（创建订单） | [order-create.md](./order-create.md) |
| 订单（透传） | GET /api/v1/orders/monthly（月订单列表） | [order-monthly.md](./order-monthly.md) |
| 订单（透传） | GET /api/v1/orders/:orderId（订单详情） | [order-detail.md](./order-detail.md) |
| 订单（透传） | GET /api/v1/orders/:orderId/payment-status（订单支付状态） | [order-payment-status.md](./order-payment-status.md) |

## 通用约定

### 统一响应结构（非订单类接口）

所有非订单类接口统一返回如下 JSON（HTTP 状态码通常为 `200`，业务结果以 `code` 判断）：

```json
{
  "code": 0,        // 业务码：0 成功；非 0 见 ../error-codes.md
  "message": "ok",  // 文案
  "data": {}        // 业务数据；失败时通常为 null
}
```

- 业务错误默认以 **HTTP 200 + 非零 `code`** 返回（见 `response.ErrorBiz` / `ErrorOf`）。
- 校验类错误（`4xxxxx`，如参数非法、未授权）会暴露真实 HTTP 状态码（400 / 401 / 403）。
- 系统错误（`1xxxxx`）文案收敛为「服务异常，请稍后重试」，不泄露内部细节。

### 鉴权

- 全局中间件链：`OTel` → `Trace` → `Logger` → `CORS` → `ExceptionInterceptor` → `Auth`。
- `Auth` 基于 `cfg.AuthWhitelist` 放行；**客户登录接口在白名单内，无需鉴权**，其余接口需携带登录态，未通过返回 `40100 ErrUnauthorized`。
- **订单类接口**额外校验请求头 `X-Customer-Id`（缺失或非法返回 `40100 ErrUnauthorized`），由 `Handler.customerID` 解析。

### 订单类接口（透传代理）

`/api/v1/orders/*` 由本服务**透传**至下游订单中台（`DownstreamAPIPrefix = /api/v1/customer/orders`）：
- 仅做 `X-Customer-Id` 鉴权与请求转发，**不套用统一 `{code, message, data}` 包装**。
- 响应为下游原始报文（HTTP 状态码与 body 直接透传），请求/响应结构由下游订单服务契约定义，本文档仅说明路由与鉴权，详细字段请参考下游订单服务文档（待补充）。

### 错误码

- 完整错误码清单（分段规范、通用错误、各模块错误）见 [../error-codes.md](../error-codes.md)。
- 各接口文档的「错误码」小节仅列出**本接口可能触发**的码值，供前端按 `code` 分支处理。
- 错误码分段：业务 `2xxxxx`、下游 `3xxxxx`、参数/校验 `4xxxxx`、限流 `5xxxxx`、系统 `1xxxxx`。
