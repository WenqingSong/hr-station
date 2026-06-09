# K9-UI-ResumeEdit-D-C-Fix：子模块全删后的空状态闪烁修复

> 日期：2026-06-09 | 性质：前端修复，未修改后端/数据库

---

## 核心变更

### 问题根因

删除最后一个子项时，`removingItemKeys.size === 0` 全局检查导致空状态被阻塞，add button 在 exiting item 高度归零后裸露显示。

### 修复策略

**三状态机**：hasItems / removingLastItem / empty，通过 `list.length === 0`（而非全局 `removingItemKeys.size === 0`）判断 empty，通过 `list.some(item => !removingItemKeys.has(key(item)))` 判断是否显示 add button。

### 具体修改

| 列表 | 修改项 | 说明 |
| -- | -- | -- |
| 教育经历 | 移除 `&& removingItemKeys.size === 0` | `educations.length === 0` 直接判断 empty |
| 教育经历 | add button 加 `hasVisibleEdu` 守卫 | `educations.some(e => !removingItemKeys.has(...))` |
| 项目经历 | 同上两处 | 同上模式 |
| 实习经历 | 同上两处 | 同上模式，空状态文案补充副标题 |
| 证书与获奖 | 重构为 IIFE 计算 `hasVisibleCerts`/`hasVisibleAwards` | add buttons 有条件渲染 |
| 作品链接 | 新增空状态卡片 | 新增 empty state |
| 作品链接 | add button 加守卫 | `works.some(w => !removingItemKeys.has(...))` |
| 外语考试 | add button 加守卫 | `languageExamDrafts.some(le => !removingItemKeys.has(...))` |

### 空状态文案更新

| 模块 | description |
| -- | -- |
| 实习经历 | 可以补充实习、兼职或校内实践经历。 |
| 证书与获奖 | 可以补充语言、云计算、软件开发等相关证书或竞赛获奖。 |
| 作品链接 | 可以补充 GitHub、博客、作品集或项目地址。 |

---

## 构建/检查

| 检查 | 结果 |
| -- | -- |
| npm build | ✅ 649 kB |
| go build | ✅ |
| go vet | ✅ |
| lint | ✅ |
| disabled terms | ✅ 仅合法代码中的 undefined |

---

## 未修改范围

- ✅ 后端/数据库/init.sql 零修改
- ✅ 未执行 Git
- ✅ 未引入新依赖
