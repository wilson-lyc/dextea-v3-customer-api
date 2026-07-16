# dextea-customer-api

基于 Go + [gin](https://github.com/gin-gonic/gin) 框架的后端服务。

## 项目结构

```
.
├── cmd/
│   └── server/
│       └── main.go          # 程序入口
├── internal/
│   ├── config/              # 配置加载（环境变量）
│   ├── handler/             # HTTP 处理器
│   └── router/              # 路由注册
├── go.mod
└── README.md
```

> `internal` 目录下的代码仅限本项目内部引用，避免被外部模块误用。

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

| 环境变量       | 默认值         | 说明           |
| -------------- | -------------- | -------------- |
| `PORT`         | `8080`         | 服务监听端口   |
| `ENVIRONMENT`  | `development`  | 运行环境       |
