# Docker 说明

## 当前状态

待编写（docker-compose.yml 存在但内容为空，Dockerfile 不存在）

## 规划服务

| 服务 | 镜像 | 端口 | 说明 |
|------|------|------|------|
| web | node:20-alpine | 5173 | 前端开发服务 |
| server | golang:1.22-alpine | 8080 | 后端 API 服务 |
| mysql | mysql:8.0 | 3306 | 数据库 |
| redis | redis:7-alpine | 6379 | 缓存 |

## docker-compose.yml 规划

```yaml
version: '3.8'
services:
  mysql:
    ...
  redis:
    ...
  server:
    ...
  web:
    ...
```

## 后续待补充

- 完整的 docker-compose.yml 内容
- 各服务 Dockerfile
- 开发/生产环境配置差异
- 数据持久化方案

## 关联文档

- [环境变量.md](./环境变量.md)
- [部署说明.md](./部署说明.md)
