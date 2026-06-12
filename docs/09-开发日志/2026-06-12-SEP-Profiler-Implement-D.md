# 2026-06-12 SEP-Profiler-Implement-D — Hy3 Resume/Job Semantic Profiler

## 本阶段目标

实现 SEP 语义证据画像层的 Hy3 Profiler，生成高质量简历/岗位语义画像并写入新表。

## 新增文件

| 文件 | 内容 |
|---|---|
| `server/internal/model/semantic_profile.go` | 新增 LLMOutput 类型 + raw code 黑名单 + sanitizer + persistence helper |
| `server/internal/service/semantic_profile_prompt.go` | Resume/Job Profiler Prompt Builder |
| `server/internal/service/semantic_profile_validator.go` | 15 条确定性校验规则 |
| `server/internal/service/semantic_profile_profiler.go` | Hy3 Profiler 主逻辑（Ensure + Generate + persist） |

## Profiler 架构

| 阶段 | 实现 |
|---|---|
| Source Extractor | `BuildSourceFromResume()` / `BuildSourceFromJob()` |
| Prompt Builder | Hy3 high system+user 中文 prompt，严格禁止 raw code |
| Hy3 Profiler | `llm.Client.Chat()` with reasoning_effort=high, temp=0.3, maxTokens=4096 |
| JSON Parse | `json.Unmarshal` + markdown fence 清理 + 括号修复 |
| Sanitizer | `SemanticProfileLLMOutput.SanitizeLLMOutput()` 黑名单扫描 |
| Validator | `ValidateSemanticProfileOutput()` 15 条规则 |
| Persistence | `SaveVerifiedProfile()` → semantic_profiles + semantic_profile_tags |

## LLM 配置

| 项目 | 值 |
|---|---|
| model_provider | hy3 |
| model_name | quality model / hy3-preview |
| reasoning_effort | high |
| temperature | 0.3 |
| max_tokens | 4096 |

## 构建验证

| 项目 | 结果 |
|---|---|
| go build | ✅ 通过 |
| go vet | ✅ 通过 |
| go test | ✅ 全部通过 |

## 未做事项

- ❌ 未实现 SemanticProfileGate
- ❌ 未接入 Fast/Deep 推荐
- ❌ 未启动全量 backfill
- ❌ 未修改旧表
- ❌ 未创建 probe CLI（Hy3 API key 需运行时注入，不适合静态 probe）
