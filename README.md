# Anvil

Anvil 是一个面向 Agent 场景的轻量化 Go 平台脚手架。

它提供一套可直接起步的工程骨架，用于搭建带有 HTTP 服务、异步任务、LLM 能力与基础中间件的 Go Agent 项目。

## 用 CLI 创建新项目

推荐使用 `anvil` CLI 生成独立项目，无需手工 clone 本仓库再整理配置。

### 本地调试 CLI

在 `anvil` 仓库根目录：

```bash
go run ./cmd/anvil new my-app
```

可选参数：

```bash
go run ./cmd/anvil new my-app --out /tmp/workspace
go run ./cmd/anvil new my-app --force
go run ./cmd/anvil new my-app --module github.com/acme/my-app
```

`my-app` 同时作为项目目录名；未指定 `--module` 时，Go module 名与目录名相同。指定 `--module` 后，目录仍为 `my-app`，`go.mod` 与 import 使用 `--module` 的值。

### 全局安装

```bash
go install github.com/ZyqAlwaysCool/anvil/cmd/anvil@latest
```

安装完成后可直接执行：

```bash
anvil new my-app
```

### 生成后启动

```bash
cd my-app
cp configs/.env.example .env
make run
```

默认配置为保守基线，不依赖 Redis、Mongo、LLM，可直接完成最小 HTTP 启动。

若生成时未指定 `--module`，也可在生成后自行修改：

```bash
go mod edit -module github.com/acme/my-app
go mod tidy
```

## 特性

- Gin HTTP 服务
- Server / Worker 双运行角色
- 统一配置加载与启动期校验
- `slog` 结构化日志与按天轮转
- 平台级错误与统一响应协议
- TraceID / AccessLog / CORS / JWT 中间件
- Redis / MySQL / SQLite / Mongo 可选装配
- Redis Stream + MongoDB 异步任务系统
- LLM 客户端封装
- Prompt 模板、结构化输出、轻量 Workflow
- 根目录 `test/` 下的完整契约测试

## 适用场景

Anvil 适合以下类型的项目起步：

- 需要 HTTP 接口和异步任务消费的 Agent 服务
- 需要将 LLM 调用封装为稳定平台能力的 Go 项目
- 希望以较小工程复杂度获得统一工程基线的团队内部服务

它的定位是“生产项目起步所需的核心工程骨架”。

## 在本仓库内直接开发

若你在 `anvil` 仓库本身开发平台能力，而不是使用 CLI 生成的新项目：

### 1. 本地最小启动

```bash
cp configs/.env.example .env
make run
```

### 2. 启用完整任务链路

```env
TASK_ENABLED=true
REDIS_ENABLED=true
MONGO_ENABLED=true
REDIS_ADDRS=127.0.0.1:6379
MONGO_URI=mongodb://127.0.0.1:27017
MONGO_DATABASE=anvil
```

```bash
make run
make run-worker
```

### 3. 使用 Docker Compose

```bash
docker compose up --build
```

扩容 Worker：

```bash
docker compose up --scale worker=2
```

说明：

- `configs/.env.example` 表示默认保守基线
- `docker-compose.yaml` 会通过 `environment:` 覆盖启用完整任务链路
- server / worker 日志挂载到 `deploy_docker/logs/`（对应容器内 `LOG_DIR`，默认 `var/log/anvil`）

## 关键能力说明

### 配置

- 平台配置唯一真源：`internal/platform/config`
- 所有平台模块通过统一 `Config` 装配
- 模块默认按需启用，不默认强依赖外部服务

### 任务系统

- 任务创建后立即返回 `task_id`
- HTTP 层只负责创建任务和查询任务
- Worker 异步消费 Redis Stream 消息
- MongoDB 持久化任务记录

### LLM

- 对外暴露平台自定义类型，不直接泄漏 SDK 类型
- 支持普通生成、真实流式生成、工具调用、JSON Schema 输出约束

## 目录结构

```text
anvil/
├── cmd/
│   ├── anvil/              # CLI：anvil new
│   ├── server/             # HTTP 入口
│   └── worker/             # Worker 入口
├── internal/
│   ├── cli/scaffold/       # CLI 实现
│   ├── platform/           # 平台能力
│   ├── server/             # 路由聚合
│   └── agent/              # 示例业务接入
├── templates/project/        # CLI 生成模板（embed）
├── prompts/
├── test/
├── codex-docs/
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

## 常用命令

```bash
make build
make test
make lint
make run
make run-worker
go run ./cmd/anvil new my-app
```