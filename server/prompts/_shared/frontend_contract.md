## 前端字段契约

所有输出字段必须兼容前端，数组不能为 null，至少输出 []。

通用字段：
- generated_by: "llm_agent" / "fast_tag_engine" / "fallback_agent"
- agent_status: "verified" / "repaired" / "success"
- fallback: true / false
- generation_tier: "fast" / "quality"
- score 范围 0-100

推荐：recommendation_score, recommendation_reason, match_highlights, risk_points
匹配：match_score, match_highlights, gap_points, risk_points
优化：optimization_score, quick/structure/content/keyword/evidence_fix_items, risk_points, rewrite_examples
对比：winner_resume_id 必须属于 from/to
