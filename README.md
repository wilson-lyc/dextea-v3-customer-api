# dextea-customer-api

基于 Go + [gin](https://github.com/gin-gonic/gin) 框架的后端服务。

## 项目结构

按**业务模块**拆包：`internal/<模块名>/` 下自带 handler / service / repository，
而非全局扁平的分层目录。

```
.
├── cmd/
│   └── server/
│       └── main.go          # 程序入口（组装各模块依赖）
├── internal/
│   ├── config/              # 配置加载（环境变量）
│   ├── db/                  # 数据库连接助手（驱动无关）
│   ├── router/              # gin 引擎与全局中间件
│   └── demo/                # Demo 业务模块
│       ├── handler.go       # 接口定义：路由注册 + gin 处理函数
│       ├── dto.go           # 接口数据提取：请求/响应结构体、绑定校验
│       ├── service.go       # 业务逻辑：编排 repository 与领域规则
│       ├── repository.go    # 数据访问：DB 连接与 SQL 操作
│       └── <未来按需补充 model/consts 等>
├── go.mod
└── README.md
```

> `internal` 目录下的代码仅限本项目内部引用，避免被外部模块误用。

### 分层职责

| 层            | 文件           | 职责                                               |
| ------------- | -------------- | -------------------------------------------------- |
| 接口定义      | `handler.go`   | 定义路由、暴露 HTTP 接口、传递上下文、写出响应     |
| 数据提取      | `dto.go`       | 请求/响应结构体、gin 绑定与校验                    |
| 业务逻辑      | `service.go`   | 编排 repository，承载业务规则，不感知 HTTP/SQL     |
| 数据访问      | `repository.go`| 专管 `sql.DB` 连接与 SQL 语句                       |

新增业务模块时，在 `internal/` 下新建同名目录并按上述四层落地，
最后在 `main.go` 中组装 `repository -> service -> handler` 并注册到路由即可。

## 快速开始

```bash
# 安装依赖
go mod tidy

# 运行服务
go run ./cmd/server

# 自定义端口 / 环境
PORT=8080 ENVIRONMENT=development go run ./cmd/server
```

## 接口

### 健康检查

```bash
curl http://localhost:8080/health
```

返回：

```json
{
  "status": "ok",
  "service": "dextea-customer-api"
}
```

## 配置项

| 环境变量         | 默认值              | 说明                                       |
| ---------------- | ------------------- | ------------------------------------------ |
| `PORT`           | `8080`              | 服务监听端口                               |
| `ENVIRONMENT`    | `development`       | 运行环境（production 时关闭 gin 调试模式） |
| `SERVICE_NAME`   | `dextea-customer-api` | 服务名（出现在健康检查响应中）           |
| `DATABASE_DRIVER`| `sqlite`            | 数据库驱动名（需自行导入对应驱动包）       |
| `DATABASE_DSN`   | （空，不启用）      | 数据库连接串；留空则 health 显示 `db:disabled` |

接入真实数据库示例（以纯 Go 的 SQLite 为例）：

```bash
# 1. 在 go.mod 中加入驱动
go get modernc.org/sqlite
# 2. 在 cmd/server/main.go 顶部匿名导入
import _ "modernc.org/sqlite"
# 3. 启动时指定 DSN
DATABASE_DSN="file:app.db?_pragma=busy_timeout=5000" go run ./cmd/server
```
