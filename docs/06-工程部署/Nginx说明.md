# Nginx 说明

> 最后更新：2026-05-30（阶段 4-3.7.B.R.3 Nginx 代理与 CORS 复核修正）

## 1. 概述

Nginx 在混合式开发环境中作为统一反向代理入口，运行在 Docker 容器中，将请求转发到宿主机上运行的 Backend / Frontend 进程。

## 2. 架构

```
浏览器 → Nginx 容器 (:80) → host.docker.internal → 宿主机进程
                                              ├→ Frontend :5173
                                              └→ Backend :8080
```

## 3. 代理规则

| 入口 | 目标 | 说明 |
|------|------|------|
| `http://localhost/` | `http://host.docker.internal:5173` | 前端页面（Vite Dev Server） |
| `http://localhost/api/` | `http://host.docker.internal:8080` | 后端 API（不 strip `/api` 前缀） |
| `http://localhost/healthz` | `http://host.docker.internal:8080/healthz` | 健康检查 |

### 关键规则

1. **`/api/` 不 strip `/api` 前缀**：后端路由以 `/api/` 开头，Nginx 直接转发不修改路径
2. **Vite HMR WebSocket 代理**：支持 `Upgrade` / `Connection` 头，确保 HMR 正常工作
3. **SSE / 长连接预留**：`/api/` 路径禁用 `proxy_buffering`，设置 `proxy_read_timeout 300s`
4. **Proxy Headers**：`Host` / `X-Real-IP` / `X-Forwarded-For` / `X-Forwarded-Proto`

## 4. host.docker.internal

Nginx 容器通过 `host.docker.internal` 访问宿主机服务：

- **macOS / Windows**：Docker Desktop 自动支持
- **Linux**：需要在 `docker-compose.yml` 中配置 `extra_hosts: host.docker.internal:host-gateway`

## 5. 配置文件

| 文件 | 说明 |
|------|------|
| `nginx/nginx.conf` | Nginx 主配置 |
| `nginx/conf.d/default.conf` | 反向代理 Server 配置 |

## 6. 日志查看

```bash
make nginx-logs    # 查看 Nginx 容器实时日志
docker compose logs nginx  # 同上
```

## 7. 开发直连模式

开发时也可不走 Nginx，直接访问宿主机进程：

- 前端：`http://localhost:5173`
- 后端：`http://localhost:8080`
- 此模式需要后端 CORS 支持（R.3 实现）

## 8. 故障排查

### Nginx 返回 502

可能原因：
1. Backend / Frontend 进程未启动：`make status` 检查
2. Vite 未绑定 `0.0.0.0`：确认 `npm run dev -- --host 0.0.0.0`
3. `host.docker.internal` 无法解析：确认 `extra_hosts` 配置

### Nginx 返回 404

可能原因：
1. 请求路径不在代理规则中
2. 后端路由不存在

### Linux 下 host.docker.internal 不可达

确认 `docker-compose.yml` 中 Nginx 服务包含：

```yaml
extra_hosts:
  - "host.docker.internal:host-gateway"
```

## 9. R.3 更新

4-3.7.B.R.3 已完成 Nginx 代理与 CORS 复核修正。当前推荐通过 `http://localhost/` 访问页面，通过 `/api/` 请求后端，由 Nginx 做同源代理。后端 CORS 已改为环境变量配置化，详见 `docs/06-工程部署/CORS说明.md`。

## 10. R.4 更新

4-3.7.B.R.4 已完成混合式一键启动集成验收。Nginx 统一入口 `/`、`/api/`、`/healthz` 均通过验收。

## 11. R.5 更新

4-3.7.B.R.5 已完成 README 与部署说明更新。Nginx 代理规则已在 README 和部署说明中写清楚。4-3.7.B.R 返工阶段已完成。
