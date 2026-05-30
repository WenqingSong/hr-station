# CORS 说明

> 最后更新：2026-05-30（阶段 4-3.7.B.R.3 Nginx 代理与 CORS 复核修正）

## 1. 背景

项目采用混合式开发环境：

```text
Docker Compose：MySQL + Nginx
Host 本机后台进程：Backend (localhost:8080) + Frontend (localhost:5173)
Nginx 统一入口：http://localhost/
```

浏览器通过 Nginx 访问前端和后端 API，正常情况下是同源请求，不触发 CORS。但开发调试时前端可能直连后端（`localhost:5173 → localhost:8080`），此时会触发跨域预检。

## 2. 当前混合式开发访问模式

```text
主推荐：
  浏览器 → http://localhost/ → Nginx → 前端/后端（同源，不触发 CORS）

后端直连调试：
  浏览器 → http://localhost:5173 → http://localhost:8080/api/...（跨域，触发 CORS）
```

## 3. Nginx 同源代理策略

当浏览器通过 `http://localhost/` 访问时，前端页面和 API 请求都发到同一个域（`http://localhost`），由 Nginx 做反向代理：

```text
/          → host.docker.internal:5173（前端 Vite Dev Server）
/api/      → host.docker.internal:8080（后端 API，不 strip /api）
/healthz   → host.docker.internal:8080/healthz
```

同源请求不触发 CORS，Nginx 代理不添加 CORS 头。

## 4. 后端直连调试策略

当浏览器直接访问 `http://localhost:5173` 的前端页面，前端请求 `http://localhost:8080/api/...` 时，触发跨域：

```text
Origin: http://localhost:5173
→ 后端 CORS 中间件检查 Origin 在 allowlist 中
→ 返回 Access-Control-Allow-Origin: http://localhost:5173
```

## 5. CORS 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `CORS_ALLOWED_ORIGINS` | `http://localhost,http://localhost:5173,http://127.0.0.1,http://127.0.0.1:5173` | 允许的 Origin 列表，逗号分隔 |
| `CORS_ALLOWED_METHODS` | `GET, POST, PUT, PATCH, DELETE, OPTIONS` | 允许的 HTTP 方法 |
| `CORS_ALLOWED_HEADERS` | `Content-Type, Authorization, X-Requested-With` | 允许的请求头 |
| `CORS_ALLOW_CREDENTIALS` | `false` | 是否允许携带凭证（Cookie） |

## 6. 开发环境配置

`.env.example` 中 CORS 默认配置适用于混合式开发环境：

```env
CORS_ALLOWED_ORIGINS=http://localhost,http://localhost:5173,http://127.0.0.1,http://127.0.0.1:5173
CORS_ALLOWED_METHODS=GET, POST, PUT, PATCH, DELETE, OPTIONS
CORS_ALLOWED_HEADERS=Content-Type, Authorization, X-Requested-With
CORS_ALLOW_CREDENTIALS=false
```

说明：
- `http://localhost` 和 `http://127.0.0.1`：Nginx 同源代理场景（无端口）
- `http://localhost:5173` 和 `http://127.0.0.1:5173`：前端直连后端调试场景
- 任何 `localhost:` 或 `127.0.0.1:` 带端口的 Origin 也自动允许（开发便利性）
- `CORS_ALLOW_CREDENTIALS=false`：当前使用 JWT Authorization Header，不使用 Cookie

## 7. 生产环境配置边界

生产环境必须：

```text
1. 只允许正式域名；
2. 不使用 *；
3. 不默认允许所有 Origin；
4. 不把 localhost 当作 production 配置；
5. 真实域名通过部署环境变量注入。
```

生产环境示例：

```env
CORS_ALLOWED_ORIGINS=https://offer-catcher.example.com
CORS_ALLOW_CREDENTIALS=false
```

## 8. OPTIONS 预检

CORS 中间件处理 OPTIONS 预检请求的逻辑：

```text
1. 读取请求 Origin 头；
2. Origin 在 allowlist 中 → 返回 204 + CORS 头；
3. Origin 不在 allowlist 中 → 不设置 CORS 头，返回 405（由业务 Handler 处理）；
4. 无 Origin 头 → 不设置 CORS 头，正常处理。
```

预检缓存时间：86400 秒（24 小时）。

## 9. Authorization / Content-Type

`CORS_ALLOWED_HEADERS` 默认包含：

```text
Content-Type    - JSON 请求体
Authorization   - JWT Bearer Token
X-Requested-With - XMLHttpRequest 标识
```

如果需要自定义请求头，在 `CORS_ALLOWED_HEADERS` 中追加。

## 10. Credentials 策略

当前策略：

```text
CORS_ALLOW_CREDENTIALS=false
```

原因：
- 当前项目使用 JWT Authorization Header 认证，不使用 Cookie
- `Allow-Credentials=false` 时，前端不需要设置 `credentials: 'include'`
- 如果后续使用 Cookie，改为 `true` 并确保 `Access-Control-Allow-Origin` 不使用 `*`

## 11. 安全边界

```text
1. 不在 allowlist 中的 Origin 不会获得 Access-Control-Allow-Origin 头
2. 浏览器会阻止跨域请求（CORS 机制保障）
3. 不会返回 Access-Control-Allow-Origin: *
4. evil.example.com 等非项目域名不会被允许
```

## 12. 验证命令

### Nginx 同源代理 OPTIONS 预检

```bash
curl -s -X OPTIONS http://localhost/api/auth/login \
  -H "Origin: http://localhost" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type,Authorization" \
  -D - -o /dev/null
```

期望：204 + `Access-Control-Allow-Origin: http://localhost`

### 后端直连 OPTIONS 预检

```bash
curl -s -X OPTIONS http://localhost:8080/api/auth/login \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type,Authorization" \
  -D - -o /dev/null
```

期望：204 + `Access-Control-Allow-Origin: http://localhost:5173`

### 不允许的 Origin

```bash
curl -s -X OPTIONS http://localhost:8080/api/auth/login \
  -H "Origin: http://evil.example.com" \
  -H "Access-Control-Request-Method: POST" \
  -D - -o /dev/null
```

期望：无 `Access-Control-Allow-Origin` 头

## 13. 当前限制

- CORS 配置通过环境变量读取，Go 后端启动时一次性加载，不支持热更新
- `localhost:` / `127.0.0.1:` 任意端口的自动允许是开发便利性设计，生产环境应去掉

## 14. 后续计划

- 生产部署时必须收紧 `CORS_ALLOWED_ORIGINS` 为正式域名
- 如使用 Cookie 认证，需将 `CORS_ALLOW_CREDENTIALS` 改为 `true`
