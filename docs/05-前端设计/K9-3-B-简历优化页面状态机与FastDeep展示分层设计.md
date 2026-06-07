# K9-3-B: 简历优化页面状态机与 Fast/Deep 展示分层设计

Date: 2026-06-07 | Author: AI Agent | Phase: K9-3-E Design

## 1. 页面状态机

```text
loading_latest → GET /api/resumes/{id}/optimization/latest
  ├── has_result=true, tier=fast → fast_done
  ├── has_result=true, tier=quality → deep_done
  └── has_result=false → fast_loading → POST fast → fast_done

fast_done:
  → click Deep → deep_loading → POST quality → deep_done
  → error → deep_error (保留 current result)
```

## 2. URL 参数解析

```text
resumeId → resume_id
jobId → job_id (optional)
targetDirection → target_direction (optional)

derive:
  has jobId → target_job
  has targetDirection → target_direction
  else → general
```

## 3. Fast 简化展示

- optimization_score + optimization_level (score badge)
- one_sentence_summary
- quick_fix_items ≤ 5 条 (section: issue → action)
- risk_points ≤ 2 条
- delivery_focus
- "生成深度优化" upsell

## 4. Deep 完整展示

- optimization_score + optimization_level
- one_sentence_summary
- quick_fix_items, structure_fix_items, content_fix_items, keyword_fix_items, evidence_fix_items (分节展示)
- risk_points (全部)
- **rewrite_examples** (视觉重点：before/after 并排 + section + reason)
- delivery_focus
- long_term_improvements

## 5. AgentStepLoadingOverlay

- Deep loading 时 overlay 居中显示
- 4 步骤：profile → analyze → gaps → suggestions
- status="loading" 自动推进

## 6. 右侧信息区

- 当前简历名称 + 画像标签
- 优化模式 + 目标方向/岗位
- Deep 按钮（根据 phase 状态变化）
