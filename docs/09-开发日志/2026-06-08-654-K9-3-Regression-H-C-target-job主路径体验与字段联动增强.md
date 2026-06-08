# K9-3-Regression-H-C — target_job 主路径体验与字段联动增强

日期：2026-06-08
性质：前端增强。1 文件修改（ResumeOptimizationPage.tsx）。

## 背景

K9-3-Regression-H-B-Frontend 完成了入口弹窗和页面骨架重构。本阶段在 H-B 基础上增强 target_job 主路径下的信息联动体验。

## 增强内容

### 1. 匹配分析 latest 联动

- 在 `resumeId + jobId` 场景下调用 `getLatestMatchAnalysis()` 读取匹配分析最新结果
- 有 match 结果：显示匹配分、匹配等级、主要缺口、主要优势、"查看匹配分析"按钮
- 无 match 结果：不阻塞页面，显示提示"尚未生成匹配分析，可先生成匹配分析获得更完整依据"
- API 失败：静默处理，显示"暂不可用"

### 2. 顶部优化对象增强

- 简历卡：名称 + 画像 + 学历/专业/学校 + 项目数/实习数
- 岗位卡：标题 + 方向/城市/招聘类型/部门 + 岗位关键词标签
- 模式+评级：优化模式 + 快速/深度 + 匹配分（如有）
- general 模式：顶部展示黄色提示栏"通用优化模式，选择目标岗位后建议更具体"

### 3. 右侧栏增强（4 张卡片）

| 卡片 | 内容 |
|------|------|
| 当前简历 | 名称 + 画像 + 学历 + 项目/实习数 + 技能标签 |
| 目标岗位 | 标题 + 方向/城市/招聘类型/部门 + 岗位关键词 |
| 匹配分析依据 | 匹配分/等级 + 缺口 + 优势 + 查看/生成匹配分析按钮 |
| 操作 | 重新选择 + 深度优化 + 匹配分析 + 岗位详情 |

### 4. DeepView 增强

- Top3：P1/P2/P3 标签 + 问题 + "建议修改位置：xxx" + 怎么改
- JD 证据区：蓝色背景区分
- 改写示例：新增"复制修改后"按钮（Clipboard API）
- `RewriteExampleCard` 独立组件

### 5. FastView 增强

- 标题改为"投递前先改 N 件事"
- 使用 Target 图标代替 FileEdit
- 优化升级提示文案更具体

### 6. 空值处理

- 所有字段使用 optional chaining 和过滤器避免 `undefined`/`null`
- 时间格式使用 `YYYY-MM-DD HH:mm`
- 不暴露 `mock_agent`/`mock_quality`

## 构建 / 测试

| 验收项 | 结果 |
|--------|------|
| npm build | ✅ 626.23 KB JS, 143.35 KB CSS |
| go build ./... | ✅ |
| go vet ./... | ✅ |
| go test ./... | 6 FAIL — pre-existing（agent/service） |
| check-encoding | ✅ OK |

## 未修改范围确认

- 未修改 K9-1 岗位推荐业务逻辑
- 未修改 K9-2 岗位匹配分析后端
- 未修改 K9-3 后端 API / Service / Agent / Repository / Model
- 未修改 K9-4 简历版本对比
- 未修改 server/sql/init.sql
- 未修改数据库结构
- 未删除真实简历
- 未执行 reset-up
- 未执行 Git 命令
