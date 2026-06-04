# K8-5-25-F — payload / adapter / 回填兼容整体验收与小修

## 性质
acceptance + minor fix（整体验收 9 模块 + payload + backfill + 1 行 import 清理）。

## 本阶段目标

对 K8-5-25-C/D/E 做整体验收，判断 payload/adapter/回填是否与 B 设计冻结一致。只做最小修正。

## 9 模块结构复核

| 模块 | 结果 |
|------|:---:|
| 1-简历基础 | ✅ ResumeBasicsSection |
| 2-基础信息 | ✅ BasicInfoSection |
| 3-教育经历 | ✅ EducationSection |
| 4-项目经历 | ✅ ProjectSection |
| 5-技能信息 | ✅ SkillsSection |
| 6-实习经历 | ✅ InternshipSection |
| 7-证书与获奖 | ✅ CertificatesSection (含 cert+award) |
| 8-作品/个人主页 | ✅ PortfolioExtraSection |
| 9-自我评价 | ✅ PortfolioExtraSection (独立) |
| JobIntentionSection | ❌ 不再渲染 ✅ |
| AwardsSection | ❌ 不再单独渲染 ✅ |

**9 模块完全匹配 K8-5-25-B ✅**

## payload 复核

| 字段 | 当前结构 | 结果 |
|------|---------|:---:|
| `schema_version` | `2` | ✅ |
| `resume_name` | resumeName 或 fallback | ✅ |
| `target_direction` | `""` | ✅ |
| `target_role` (V1) | `""` | ✅ |
| `job_intention` | 不传 | ✅ |
| `profile` | name/phone/email/currentCity + school/degree/major/graduation_year from educations[0] | ✅ |
| `educations` | V2 13字段 | ✅ |
| `internships` | V2 8字段 | ✅ |
| `projects` | V2 11字段 | ✅ |
| `skills` (V2) | `{language_exam, development_languages, ai_application}` | ✅ |
| `certificates` | V2 5字段 | ✅ |
| `awards` | V2 5字段 | ✅ |
| `portfolio_links` | `{works, homepage}` | ✅ |
| `extra_info` | `{self_evaluation, additional_info}` | ✅ |

## adapter / 回填兼容复核

| 旧结构 | 新 UI | 结果 |
|---------|-------|:---:|
| V1 skills[] | developmentLanguages | ✅ fallback |
| V2 skills 10-cat | developmentLanguages + aiApplication | ✅ |
| old portfolio github/gitee/cnb | works items | ✅ |
| old portfolio blog | homepage | ✅ |
| old demo_links/portfolio_links | works[] | ✅ |
| certs/awards | 直接回填 | ✅ |
| selfEvaluation/extra | 直接回填 | ✅ |
| job_intention | 不回填 UI | ✅ |

## 旧字段残留复核

| 字段 | 位置 | 是否残留 |
|------|------|:---:|
| github/gitee/cnb | L491-493 backfill helper | ❌ 兼容 |
| cnb.cool/github in placeholders | L1187/1394/1399 | ❌ 示例文本 |

**0 个旧字段 UI 残留 ✅**

## 小修清单

| 文件 | 修正 |
|------|------|
| `MyResumesPage.tsx` L46 | 删除未使用的 `AwardsSection` import |

## 上传能力禁止 ✅ | 误改范围 ✅

## 构建 / API / 测试

| 验收项 | 结果 |
|--------|:---:|
| `npm run build` | ✅ 2.48s |
| `go build/vet/test ./...` | ✅ |
| API 200/200/200/401 | ✅ |

## 未修改范围 ✅ | 未执行 Git ✅

## 风险与边界

```text
真实登录态端到端回归 → K8-5-25-G
Agent/Prompt/Schema 再设计 → K8-5-26-A (暂停)
```
