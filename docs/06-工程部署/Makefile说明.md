# Makefile 说明

## 当前状态

待创建（阶段 1-2 执行）

## 设计目标

1. 云环境重启后快速恢复开发环境
2. 本地开发命令统一入口
3. 前端、后端、Docker 服务分开控制
4. 后续阶段持续扩展
5. 新模型 / 新开发者快速接手

## 命令规划

### 环境管理命令

| 命令 | 目标 | 只读 | 有副作用 | 适合重启后 | 说明 |
|------|------|------|---------|-----------|------|
| `make help` | 查看所有命令 | ✅ | ❌ | ✅ | 列出所有 make 目标和说明 |
| `make doctor` | 只读检查开发环境 | ✅ | ❌ | ✅ | 检查 Go/Node/Docker/端口/依赖 |
| `make init` | 初始化开发环境 | ❌ | ⚠️ | ✅ | 复制 .env、安装依赖（幂等） |
| `make status` | 查看容器状态 | ✅ | ❌ | ✅ | docker-compose ps |

### Docker 服务命令

| 命令 | 目标 | 只读 | 有副作用 | 适合重启后 | 说明 |
|------|------|------|---------|-----------|------|
| `make up` | 启动 Docker Compose 服务 | ❌ | ⚠️ | ✅ | 启动 MySQL |
| `make down` | 停止服务（不删卷） | ❌ | ⚠️ | ❌ | docker-compose down |
| `make restart` | 重启服务 | ❌ | ⚠️ | ✅ | docker-compose restart |
| `make logs` | 查看全部服务日志 | ✅ | ❌ | ✅ | docker-compose logs -f |
| `make logs-mysql` | 查看 MySQL 日志 | ✅ | ❌ | ✅ | docker-compose logs -f mysql |
| `make reset` | ⚠️ 危险：停止并清理容器卷 | ❌ | 🔴 | ❌ | docker-compose down -v（删除数据） |

### 开发命令

| 命令 | 目标 | 只读 | 有副作用 | 适合重启后 | 说明 |
|------|------|------|---------|-----------|------|
| `make dev-web` | 启动前端开发服务 | ❌ | ⚠️ | ✅ | cd web && npm run dev |
| `make dev-server` | 启动后端开发服务 | ❌ | ⚠️ | ✅ | cd server && go run cmd/server/main.go |
| `make dev` | 输出开发启动提示 | ✅ | ❌ | ✅ | 打印开发环境启动步骤 |

### 构建与测试命令

| 命令 | 目标 | 只读 | 有副作用 | 适合重启后 | 说明 |
|------|------|------|---------|-----------|------|
| `make build-web` | 构建前端 | ❌ | ⚠️ | ❌ | cd web && npm run build |
| `make build-server` | 构建后端 | ❌ | ⚠️ | ❌ | cd server && go build -o bin/server cmd/server/main.go |
| `make build` | 构建前后端 | ❌ | ⚠️ | ❌ | build-web + build-server |
| `make test-server` | 运行后端测试 | ✅ | ❌ | ✅ | cd server && go test ./... |
| `make check` | 执行基础检查 | ✅ | ❌ | ✅ | doctor + test-server |
| `make clean` | 清理构建产物 | ❌ | ⚠️ | ❌ | 删除 bin/、dist/ 等 |

## 关键规则

1. `make doctor` 只检查，不修改任何文件
2. `make init` 幂等：`.env` 已存在则跳过，依赖已安装则跳过
3. `make up` 只启动 Docker Compose
4. `make reset` 必须在说明中标记为危险命令，执行前需确认
5. 不把真实密钥写进 Makefile
6. 命令尽量幂等，重复执行不破坏环境
7. 命令失败时输出清晰提示

## `make doctor` 检查项

```
✓ Go 是否安装（go version）
✓ Node.js 是否安装（node --version）
✓ npm 是否安装（npm --version）
✓ docker 是否安装（docker --version）
✓ docker-compose 是否安装（docker-compose --version）
✓ .env 是否存在
✓ node_modules 是否存在
✓ go.mod 是否存在
✓ MySQL 容器是否运行
✓ 端口 3306 是否可用
✓ 端口 5173 是否可用
✓ 端口 8080 是否可用
```

## `make init` 执行项

```
1. 如果 .env 不存在，从 .env.example 复制
2. 如果 node_modules 不存在，执行 npm install
3. 如果 go.mod 不存在，执行 go mod init
4. 输出初始化完成提示
```

## 后续扩展

| 阶段 | 新增命令 |
|------|---------|
| 阶段 2 | `make migrate`、`make seed` |
| 阶段 3 | 无 |
| 阶段 7 | `make deploy`、`make demo` |

## 关联文档

- [阶段1工程初始化方案.md](./阶段1工程初始化方案.md)
- [Docker说明.md](./Docker说明.md)
- [环境变量.md](./环境变量.md)
- [部署说明.md](./部署说明.md)
