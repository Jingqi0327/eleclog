# Eleclog

Eleclog 是一个起初专为高校宿舍场景设计的电费监控与告警系统。它能够通过定时任务调用第三方 API 抓取电费余额数据，记录用电情况，并在余额不足时自动发送邮件告警。
此外，该系统具备良好的通用性和扩展性——任何依赖第三方 API 查询电费的学校或园区，只需对接相应的 API 即可使用本系统。系统内置了多用户管理与 RBAC 权限控制，并自带一个配套的可视化管理后台页面。

## 🎯 主要功能

- **多模式运行**: 支持单体应用（Full/Main）或微服务模式（Backend/Worker/Proxy）分别部署。
- **电量数据抓取**: 基于定时任务定期获取第三方或指定代理接口的电费数据。
- **低电量预警**: 定期检测账户/房间电费余额，若低于阈值则通过邮件自动发送告警通知。
- **用户管理与 RBAC 控制**: 完善的用户注册、认证（JWT/Token）、权限校验与管理功能。
- **API 请求限流**: 基于 Redis 令牌桶算法，防止接口遭到恶意请求。
- **监控与可观测性**: 导出 Prometheus Metrics 等相关指标数据以供监控。
- **容器化部署**: 支持 Docker、Docker Compose 以及 Kubernetes(K8s) 部署。

## 🛠 技术栈

- **后端**: [Go](https://go.dev/) (>=1.20)
- **Web 框架**: [Gin](https://gin-gonic.com/)
- **RPC 通信**: [gRPC](https://grpc.io/)
- **数据库**: [PostgreSQL](https://www.postgresql.org/)
- **ORM & 数据库迁移**: [sqlc](https://sqlc.dev/), [golang-migrate](https://github.com/golang-migrate/migrate)
- **缓存 & 任务队列**: [Redis](https://redis.io/), [Asynq](https://github.com/hibiken/asynq)
- **前端**: HTML/JS (作为配套的简易管理后台，可通过本地 HTTP 服务运行)

## 🚀 快速开始

### 1. 环境准备

确保您的本地或服务器已安装以下组件：
- [Go](https://go.dev/doc/install) 
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- [Make](https://www.gnu.org/software/make/) (可选)

### 2. 配置环境变量

项目根目录包含一个示例环境变量文件 `app.env.example`。
复制一份作为 `app.env` 并根据实际情况修改配置（如数据库地址、Redis地址、邮件 SMTP 密码等）：
```bash
cp app.env.example app.env
```

### 3. 运行基础服务 (数据库与 Redis)

推荐使用 Docker Compose 快速启动依赖服务：
```bash
docker-compose up -d postgres redis
```
*(注意：具体服务名可能需要参考 `docker-compose.yaml` 进行调整)*

### 4. 数据库迁移

确保数据库已启动，并使用 `make` 命令或 `migrate` 工具执行数据库迁移：
```bash
make migrateup
```

### 5. 启动后端服务

可以使用 `make` 命令快速启动后端服务（默认包含 API, Worker, Proxy）：
```bash
make server
```

### 6. 启动前端页面

如果您想查看系统的前端界面：
```bash
make frontend
```
随后访问 `http://localhost:3001` 即可进入前端系统。

## ⚙️ 配置文件说明 (`app.env`)

关键配置项含义如下：
- `RUN_MODE`: 运行模式，可选 `full` (全部启动), `backend` (仅API), `worker` (仅任务调度/处理), `proxy` (仅RPC代理)。
- `DB_SOURCE`: PostgreSQL 数据库连接 DSN。
- `REDIS_ADDR`: Redis 连接地址。
- `HTTP_SERVER_ADDRESS` & `GRPC_SERVER_ADDRESS`: API 和 RPC 监听地址。
- `EMAIL_SENDER_*`: 用于发送低电量告警的 SMTP 发件人配置。
- `FETCH_SURPLUS_CRON` & `DETECT_LOW_BALANCE_CRON`: 抓取数据及低电量检测的 Cron 表达式。

## 📦 部署

本项目包含 `Dockerfile` 以及 `k8s/` 目录，可以轻松部署到 Kubernetes 集群。
如需构建多架构镜像，可使用：
```bash
make image
```
