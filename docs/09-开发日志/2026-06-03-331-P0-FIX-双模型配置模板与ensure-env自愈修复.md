# P0-FIX：双模型配置模板与 ensure-env 自愈修复

**日期：** 2026-06-03
**严重等级：** P0（已修复）
**性质：** config template fix + self-healing logic
**范围：** `.env.example` + `ensure_env.sh` + `.env` 安全补齐

---

## 1. P0 根因（P0-DIAG 确诊）

K7-2 实现双模型分层时：
- ✅ 修改了 `config.go`（双模型变量读取 + coalesceModel fallback）
- ✅ 修改了 `main.go`（fastRunner + qualityRunner + Agent 路由）
- ✅ 修改了运行时 `.env`（添加 LLM_FAST_MODEL / LLM_QUALITY_MODEL / LLM_DEFAULT_TIER）
- ❌ **遗漏了同步 `.env.example`（模板文件）**

后果：云原生环境 `.env` 丢失后，`ensure-env` 从 `.env.example` 复制生成单模型 `.env`，双模型分层失效。

---

## 2. `.env.example` 修复内容

| 变更项 | 旧值 | 新值 |
|--------|------|------|
| LLM_PROVIDER | `hunyuan` | `deepseek` |
| LLM_MODEL | `hy3-preview` | `deepseek-v4-flash` |
| LLM_FAST_MODEL | 不存在 | `deepseek-v4-flash` |
| LLM_QUALITY_MODEL | 不存在 | `deepseek-v4-pro-202606` |
| LLM_DEFAULT_TIER | 不存在 | `fast` |
| LLM 区域注释 | 单模型说明 | 双模型分层设计说明 + 兼容策略 |

---

## 3. `ensure_env.sh` 自愈逻辑

新增 `ensure_env_var()` 幂等追加函数：

```bash
ensure_env_var "LLM_FAST_MODEL" "deepseek-v4-flash"
ensure_env_var "LLM_QUALITY_MODEL" "deepseek-v4-pro-202606"
ensure_env_var "LLM_DEFAULT_TIER" "fast"
```

行为：
- `.env` 不存在 → 从 `.env.example` 复制后，再确保双模型变量存在
- `.env` 已存在 → 不覆盖，只追加缺失的双模型变量
- 重复执行 → 不重复追加（幂等）

---

## 4. 当前 `.env` 安全补齐

执行 `ensure-env` 后，当前 `.env` 已补齐 3 个变量：

```
LLM_FAST_MODEL=deepseek-v4-flash       ← 新增
LLM_QUALITY_MODEL=deepseek-v4-pro-202606  ← 新增
LLM_DEFAULT_TIER=fast                   ← 新增
```

未修改任何敏感字段（LLM_API_KEY、DB_PASSWORD、JWT_SECRET、MYSQL_ROOT_PASSWORD 均未触碰）。

---

## 5. 验收结果

### 配置验收

| 文件 | 变量 | 结果 |
|------|------|:--:|
| `.env.example` | LLM_PROVIDER=deepseek | ✅ |
| `.env.example` | LLM_MODEL=deepseek-v4-flash | ✅ |
| `.env.example` | LLM_FAST_MODEL=deepseek-v4-flash | ✅ |
| `.env.example` | LLM_QUALITY_MODEL=deepseek-v4-pro-202606 | ✅ |
| `.env.example` | LLM_DEFAULT_TIER=fast | ✅ |
| `.env` | LLM_FAST_MODEL=deepseek-v4-flash | ✅ |
| `.env` | LLM_QUALITY_MODEL=deepseek-v4-pro-202606 | ✅ |
| `.env` | LLM_DEFAULT_TIER=fast | ✅ |

### 幂等性验收

```bash
bash ensure_env.sh  # 第 2 次执行
grep -c LLM_FAST_MODEL .env     → 1 ✅
grep -c LLM_QUALITY_MODEL .env  → 1 ✅
grep -c LLM_DEFAULT_TIER .env   → 1 ✅
```

### 测试结果

| 命令 | 结果 |
|------|:--:|
| `go test ./internal/... -count=1` | ALL PASS (10 packages) ✅ |
| `go build ./cmd/server` | SUCCESS ✅ |

---

## 6. 对四核心功能保护

修复后配置链路恢复正常：

| 核心功能 | 模型层级 | 模型 | Runner |
|----------|----------|------|--------|
| 岗位智能推荐 | Fast | `deepseek-v4-flash` | `fastRunner` |
| 岗位匹配分析 | Quality | `deepseek-v4-pro-202606` | `qualityRunner` |
| 简历优化建议 | Quality | `deepseek-v4-pro-202606` | `qualityRunner` |
| 简历版本对比 | Quality | `deepseek-v4-pro-202606` | `qualityRunner` |

---

## 7. 安全确认

```text
未输出真实密钥。✅
未修改 LLM_API_KEY。✅
未修改 DB_PASSWORD。✅
未修改 JWT_SECRET。✅
未修改 MYSQL_ROOT_PASSWORD。✅
未修改后端双模型路由逻辑。✅
```

---

## 8. 修改文件清单

| 文件 | 修改内容 | 原因 | 属于 P0-FIX |
|------|----------|------|:--:|
| `.env.example` | 补齐 LLM_FAST_MODEL / LLM_QUALITY_MODEL / LLM_DEFAULT_TIER + 更新注释 | 双模型模板缺失 | 是 |
| `scripts/dev/ensure_env.sh` | 新增 `ensure_env_var()` 自愈函数 | .env 缺失双模型变量时自动补齐 | 是 |
| `.env` | 安全追加 3 个双模型变量（未改任何敏感字段） | 当前 .env 补齐 | 是 |

---

## 9. 本次未执行 Git 命令说明

本次未执行任何 Git 命令，也未提交代码。
