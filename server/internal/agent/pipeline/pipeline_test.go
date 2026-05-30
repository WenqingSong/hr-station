package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"offer-catcher/server/internal/agent/repair"
	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// ---------------------------------------------------------------------------
// fakeClient: stub llm.Client for unit tests
// ---------------------------------------------------------------------------

type fakeClient struct {
	calls     int
	responses []string // responses[0] = first call, responses[1] = repair call, etc.
	err       error
	index     int // tracks which response to return
}

func (f *fakeClient) Chat(_ context.Context, _ llm.ChatRequest) (*llm.ChatResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	resp := ""
	if f.index < len(f.responses) {
		resp = f.responses[f.index]
		f.index++
	} else if len(f.responses) > 0 {
		resp = f.responses[len(f.responses)-1]
	}
	return &llm.ChatResponse{
		Provider:        llm.ProviderNoop,
		Model:           "fake-model",
		RawContent:      resp,
		FinishReason:    "stop",
		RawOutputLength: len(resp),
		DurationMs:      42,
	}, nil
}

func (f *fakeClient) Provider() llm.Provider { return llm.ProviderNoop }

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestRunner(fc *fakeClient) *Runner {
	promptDir, _ := filepath.Abs("../../../prompts")
	schemaDir, _ := filepath.Abs("../../../prompts/schema")

	pl := prompt.NewLoader(promptDir)
	sl := schema.NewLoader(schemaDir)
	v := schema.NewValidator()
	repairAgent := repair.NewAgent(fc, pl, sl, v)

	return NewRunner(fc, pl, sl, v, repairAgent)
}

func loadSchemaJSON(t *testing.T, relPath string) string {
	t.Helper()
	absPath, err := filepath.Abs("../../../" + relPath)
	if err != nil {
		t.Fatalf("failed to resolve path %s: %v", relPath, err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("failed to load schema file %s: %v", absPath, err)
	}
	return string(data)
}

// minimalMatchAnalysis returns a JSON object that conforms to
// match_analysis_schema.json.
func minimalMatchAnalysis() map[string]any {
	return map[string]any{
		"match_score":         75,
		"match_level":         "potential_match",
		"match_label":         "建议尝试",
		"screening_result":    "likely",
		"offer_go_summary":    "你的简历与该岗位有一定匹配度，建议优化后尝试投递。",
		"jd_key_requirements": []any{
			map[string]any{"requirement": "岗位要求Go相关经验", "importance": "core", "source": "requirements", "reason": "该岗位核心技能要求Go语言"},
			map[string]any{"requirement": "岗位要求数据库经验", "importance": "important", "source": "description", "reason": "后端开发需要数据库能力"},
			map[string]any{"requirement": "岗位要求团队协作能力", "importance": "optional", "source": "description", "reason": "团队协作是软技能加分项"},
		},
		"resume_evidence_summary": []any{
			map[string]any{"evidence": "简历中有Go项目经验", "related_requirement": "Go相关经验", "strength_level": "strong"},
			map[string]any{"evidence": "简历中有MySQL使用经验", "related_requirement": "数据库经验", "strength_level": "medium"},
		},
		"evidence_chain": []any{
			map[string]any{"jd_requirement": "岗位要求Go相关经验", "resume_evidence": "简历中有Go项目经验", "judgement": "matched", "explanation": "简历中有明确的Go项目经验描述"},
			map[string]any{"jd_requirement": "岗位要求数据库经验", "resume_evidence": "简历中有MySQL使用经验", "judgement": "partial", "explanation": "有数据库使用经验但深度不足"},
			map[string]any{"jd_requirement": "岗位要求微服务经验", "resume_evidence": "简历中暂未看到直接证据", "judgement": "missing", "explanation": "简历中没有微服务相关描述"},
		},
		"matched_points": []any{
			map[string]any{"point": "Go项目经验匹配", "why_it_matters": "岗位核心要求Go语言", "evidence": "简历中有Go项目经验"},
			map[string]any{"point": "数据库使用经验", "why_it_matters": "后端开发需要数据库能力", "evidence": "简历中有MySQL经验"},
		},
		"gap_points": []any{
			map[string]any{"gap": "缺少微服务经验", "impact": "可能影响技术面试表现", "how_to_improve": "建议学习微服务框架并完成实践项目"},
			map[string]any{"gap": "项目经历描述不够具体", "impact": "可能影响初筛通过率", "how_to_improve": "建议补充具体的技术细节和成果数据"},
		},
		"risk_points": []any{
			map[string]any{"risk": "缺少微服务经验", "severity": "medium", "reason": "岗位可能涉及微服务架构", "mitigation": "通过在线课程补充微服务知识"},
			map[string]any{"risk": "项目经历描述偏简单", "severity": "low", "reason": "初筛可能无法体现技术深度", "mitigation": "优化项目描述，增加技术细节"},
		},
		"resume_optimization": map[string]any{
			"p0": []any{
				map[string]any{"problem": "项目描述缺少量化数据", "why_it_matters": "量化数据能让初筛更有说服力", "rewrite_example": "如果该项目确实有相关数据，可以补充具体的性能提升比例", "action": "回顾项目经历，补充可量化的成果"},
			},
			"p1": []any{
				map[string]any{"problem": "技能栏缺少微服务相关关键词", "why_it_matters": "岗位可能通过关键词筛选简历", "rewrite_example": "如果有微服务学习经历，可以在技能栏中添加相关关键词", "action": "学习并在技能栏中添加微服务相关技能"},
			},
			"p2": []any{},
		},
		"interview_questions": []any{
			map[string]any{"question": "请介绍一下你做过的Go项目", "why_ask": "考察Go语言实际使用能力", "how_to_prepare": "准备项目的技术架构和关键决策说明", "related_jd_requirement": "Go相关经验"},
			map[string]any{"question": "你对微服务有什么了解", "why_ask": "岗位可能涉及微服务架构", "how_to_prepare": "学习微服务基本概念和常见框架", "related_jd_requirement": "微服务经验"},
			map[string]any{"question": "你在项目中如何使用数据库", "why_ask": "考察数据库使用能力", "how_to_prepare": "准备数据库设计和优化的案例", "related_jd_requirement": "数据库经验"},
		},
		"authenticity_notice": "以上建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。建议通过真实实践来积累经验。",
		"next_actions": []any{
			map[string]any{"action": "优化项目描述，补充量化成果", "priority": "p0", "expected_effect": "提升初筛通过率和技术面试表现"},
			map[string]any{"action": "学习微服务基础知识", "priority": "p1", "expected_effect": "增加技术面试表现和竞争力"},
			map[string]any{"action": "在技能栏中添加微服务关键词", "priority": "p2", "expected_effect": "提升关键词匹配度和筛选通过率"},
		},
		"analysis_basis": map[string]any{
			"job_description_used":  true,
			"job_requirements_used": true,
			"resume_used":           true,
			"data_quality":          "full_jd",
		},
		"generated_by": "llm_agent",
	}
}

func minimalRecommendation() map[string]any {
	return map[string]any{
		"strategy_summary": "根据你的简历和目标岗位，建议优先投递与Go后端相关的岗位，同时可以了解其他方向的岗位信息以拓宽选择。",
		"recommendations": []any{
			map[string]any{
				"job_id":           1,
				"title":            "Go后端开发工程师",
				"department":       "技术部",
				"city":             "深圳",
				"category":         "后端开发",
				"job_type":         "校招",
				"external_source":  "local_mock",
				"external_id":      "",
				"data_quality":     "full_jd",
				"source_url":       "",
				"recommend_level":  "priority_apply",
				"recommend_label":  "优先投递",
				"match_score":      88,
				"offer_go_summary": "你的Go技能与该岗位高度匹配，建议优先投递试试看。",
				"recommend_reason": "你的Go项目经验与岗位核心要求高度匹配，数据库经验也是加分项，整体匹配度较高。",
				"evidence": []any{
					map[string]any{"job_requirement": "Go相关经验", "resume_evidence": "简历中有Go项目经验", "judgement": "matched", "explanation": "简历中有明确的Go项目经验描述"},
					map[string]any{"job_requirement": "数据库经验", "resume_evidence": "简历中有MySQL使用经验", "judgement": "matched", "explanation": "岗位要求数据库能力，简历中有相关经验"},
				},
				"risk_points": []any{
					map[string]any{"risk": "缺少微服务经验", "severity": "medium", "mitigation": "建议在面试前学习微服务基础知识"},
				},
				"query_suggestions": []any{
					map[string]any{"query": "Go微服务开发入门教程", "why_search": "了解微服务架构的基本概念和Go语言实现"},
					map[string]any{"query": "Go后端面试常见问题", "why_search": "准备面试可能遇到的技术问题"},
				},
				"next_action": map[string]any{
					"action":          "优化简历中Go项目的描述，补充量化成果",
					"priority":        "p0",
					"expected_effect": "提升初筛通过率和技术面试表现",
				},
			},
		},
		"overall_risks": []any{
			map[string]any{"risk": "简历项目描述偏简单", "severity": "medium", "suggestion": "补充项目的技术细节和量化成果来提升竞争力"},
		},
		"resume_focus": []any{
			map[string]any{"focus": "Go项目经验描述优化", "why_it_matters": "Go技能是目标岗位的核心要求", "suggested_action": "补充Go项目的技术架构和关键成果数据"},
			map[string]any{"focus": "数据库技能展示", "why_it_matters": "数据库能力是后端开发的重要技能", "suggested_action": "在技能栏和项目描述中突出数据库使用经验"},
		},
		"authenticity_notice": "以上建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。建议通过真实实践来积累经验。",
		"analysis_basis": map[string]any{
			"resume_used":          true,
			"job_candidates_count": 5,
			"full_jd_count":        3,
			"summary_only_count":   2,
		},
		"generated_by": "llm_agent",
	}
}

func minimalResumeOptimization() map[string]any {
	return map[string]any{
		"optimization_summary": "针对目标岗位，建议重点优化Go项目描述和技能栏关键词，提升初筛通过率。",
		"offer_go_summary":    "你的Go技能与目标岗位匹配，优化项目描述后投递成功率会更高，建议优先改善量化表达。",
		"target_job": map[string]any{
			"job_id":       1,
			"title":        "Go后端开发工程师",
			"department":   "技术部",
			"category":     "后端开发",
			"data_quality": "full_jd",
		},
		"screening_focus": []any{
			map[string]any{"focus": "Go项目经验深度", "why_it_matters": "岗位核心要求Go语言能力", "related_jd_requirement": "Go相关开发经验"},
		},
		"priority_suggestions": map[string]any{
			"p0": []any{
				map[string]any{
					"issue":                   "Go项目描述缺少量化数据",
					"related_jd_requirement":  "Go相关开发经验",
					"current_resume_evidence": "简历中有Go项目经历但描述偏简单",
					"why_it_matters":          "量化数据能让初筛更有说服力",
					"optimization_goal":       "补充可量化的项目成果和技术细节",
					"rewrite_example":         "如果该项目确实有相关数据，可以补充具体的性能提升比例",
					"action":                  "回顾项目经历，补充可量化的成果数据",
					"authenticity_rule":       "不要编造具体的数字或指标，只补充真实存在的数据",
				},
			},
			"p1": []any{
				map[string]any{
					"issue":                   "技能栏缺少微服务相关关键词",
					"related_jd_requirement":  "微服务架构经验",
					"current_resume_evidence": "简历中暂未看到直接证据",
					"why_it_matters":          "岗位可能通过关键词筛选简历",
					"optimization_goal":       "在技能栏中添加微服务相关技能关键词",
					"rewrite_example":         "如果有微服务学习经历，可以在技能栏中添加相关关键词",
					"action":                  "学习微服务基础并在技能栏中添加相关技能",
					"authenticity_rule":       "只添加真正掌握或正在学习的技能，不要虚构能力",
				},
			},
			"p2": []any{},
		},
		"keyword_alignment": []any{
			map[string]any{"keyword": "Go", "status": "present", "resume_evidence": "简历中有Go项目经验描述", "suggested_action": "在项目描述中进一步突出Go技术栈的使用深度"},
			map[string]any{"keyword": "MySQL", "status": "present", "resume_evidence": "简历中有MySQL使用经验", "suggested_action": "在技能栏中明确标注MySQL版本和使用场景"},
			map[string]any{"keyword": "微服务", "status": "missing", "resume_evidence": "简历中暂未看到直接证据", "suggested_action": "学习微服务基础后添加到技能栏"},
		},
		"project_optimization": []any{
			map[string]any{
				"project_name":         "Go Web项目",
				"current_problem":      "项目描述偏简单，缺少技术细节和量化数据",
				"targeted_improvement": "补充技术架构描述和可量化的性能数据",
				"jd_relevance":         "该岗位核心要求Go语言开发能力，项目描述需要体现深度",
				"rewrite_direction":    "如果该项目确实有相关数据，可以补充QPS和响应时间等指标",
				"interview_risk":       "面试官可能会追问项目的技术细节和性能优化方案",
			},
		},
		"skill_section_optimization": []any{
			map[string]any{
				"skill":                 "Go",
				"current_status":        "present",
				"suggested_positioning":  "将Go技能放在技能栏首位，突出实际项目使用经验",
				"evidence_needed":        "在项目描述中补充Go语言的具体使用场景和成果",
			},
		},
		"summary_optimization": []any{
			map[string]any{
				"current_problem":   "简历摘要偏简单，没有突出核心能力",
				"rewrite_direction": "突出Go后端开发经验和数据库能力",
				"example":           "具有Go后端开发经验的工程师，熟悉MySQL数据库，正在学习微服务架构",
			},
		},
		"missing_evidence_plan": []any{
			map[string]any{
				"missing_evidence":     "微服务相关经验",
				"why_needed":           "目标岗位可能涉及微服务架构设计",
				"realistic_way_to_build": "通过在线课程学习微服务基础，完成一个微服务实践项目",
				"should_not_fake":      true,
			},
		},
		"rewrite_examples": []any{
			map[string]any{
				"section":                    "项目经验",
				"before_problem":             "项目描述缺少量化数据",
				"after_example":              "如果该项目确实有相关数据，可以补充具体的QPS和响应时间等指标",
				"can_use_directly":           false,
				"requires_user_verification": true,
				"verification_note":          "需要用户确认项目中的具体性能数据是否真实",
			},
			map[string]any{
				"section":                    "技能栏",
				"before_problem":             "缺少微服务相关关键词",
				"after_example":              "学习微服务后添加gRPC和微服务相关技能到技能栏",
				"can_use_directly":           false,
				"requires_user_verification": true,
				"verification_note":          "需要用户确认是否真正掌握了微服务相关技能",
			},
		},
		"next_actions": []any{
			map[string]any{"action": "优化Go项目描述，补充量化成果和技术细节", "priority": "p0", "expected_effect": "提升初筛通过率和技术面试表现"},
			map[string]any{"action": "学习微服务基础知识并完成实践项目", "priority": "p1", "expected_effect": "增加技术面试竞争力和岗位匹配度"},
			map[string]any{"action": "在技能栏中添加微服务和gRPC关键词", "priority": "p2", "expected_effect": "提升关键词匹配度和筛选通过率"},
		},
		"authenticity_notice": "以上建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。建议通过真实实践来积累经验。",
		"analysis_basis": map[string]any{
			"resume_used":           true,
			"job_description_used":  true,
			"job_requirements_used": true,
			"match_analysis_used":   false,
			"data_quality":          "full_jd",
		},
		"generated_by": "llm_agent",
	}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestPipeline_MatchAnalysis_FirstSuccess(t *testing.T) {
	validJSON := toJSON(minimalMatchAnalysis())
	fc := &fakeClient{responses: []string{validJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusSuccess {
		t.Fatalf("expected success, got: %s (error: %s: %s)", result.AgentStatus, result.ErrorCode, result.ErrorMessage)
	}
	if result.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got %s", result.GeneratedBy)
	}
	if result.RepairApplied {
		t.Error("expected RepairApplied=false on first success")
	}
	if result.FallbackRequired {
		t.Error("expected FallbackRequired=false")
	}
	if result.Parsed == nil {
		t.Error("expected non-nil Parsed")
	}
	if result.RawOutputLength <= 0 {
		t.Error("expected RawOutputLength > 0")
	}
	if fc.calls != 1 {
		t.Errorf("expected exactly 1 LLM call, got %d", fc.calls)
	}
}

func TestPipeline_Recommendation_FirstSuccess(t *testing.T) {
	validJSON := toJSON(minimalRecommendation())
	fc := &fakeClient{responses: []string{validJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "RecommendationAgent",
		PromptName:       prompt.PromptRecommendation,
		SchemaName:       schema.SchemaRecommendation,
		TargetOutputType: TargetRecommendation,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusSuccess {
		t.Fatalf("expected success, got: %s (error: %s: %s)", result.AgentStatus, result.ErrorCode, result.ErrorMessage)
	}
	if fc.calls != 1 {
		t.Errorf("expected exactly 1 LLM call, got %d", fc.calls)
	}
}

func TestPipeline_ResumeOptimization_FirstSuccess(t *testing.T) {
	validJSON := toJSON(minimalResumeOptimization())
	fc := &fakeClient{responses: []string{validJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "ResumeOptimizationAgent",
		PromptName:       prompt.PromptResumeOptimization,
		SchemaName:       schema.SchemaResumeOptimization,
		TargetOutputType: TargetResumeOptimization,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusSuccess {
		t.Fatalf("expected success, got: %s (error: %s: %s)", result.AgentStatus, result.ErrorCode, result.ErrorMessage)
	}
	if fc.calls != 1 {
		t.Errorf("expected exactly 1 LLM call, got %d", fc.calls)
	}
}

func TestPipeline_ParseFailed_RepairSuccess(t *testing.T) {
	// First call returns broken JSON, second call (repair) returns valid JSON.
	brokenJSON := `{"match_score": broken`
	validJSON := toJSON(minimalMatchAnalysis())
	fc := &fakeClient{responses: []string{brokenJSON, validJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusRepaired {
		t.Fatalf("expected repaired, got: %s (error: %s: %s)", result.AgentStatus, result.ErrorCode, result.ErrorMessage)
	}
	if result.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got %s", result.GeneratedBy)
	}
	if !result.RepairApplied {
		t.Error("expected RepairApplied=true")
	}
	if !result.RepairSuccess {
		t.Error("expected RepairSuccess=true")
	}
	if result.FallbackRequired {
		t.Error("expected FallbackRequired=false")
	}
	if result.Parsed == nil {
		t.Error("expected non-nil Parsed")
	}
	if fc.calls != 2 {
		t.Errorf("expected 2 LLM calls (initial + repair), got %d", fc.calls)
	}
}

func TestPipeline_SchemaValidateFailed_RepairSuccess(t *testing.T) {
	// First call returns JSON that parses but fails schema, second call (repair) returns valid JSON.
	invalidObj := map[string]any{"wrong_field": "value"}
	validJSON := toJSON(minimalMatchAnalysis())
	fc := &fakeClient{responses: []string{toJSON(invalidObj), validJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusRepaired {
		t.Fatalf("expected repaired, got: %s (error: %s: %s)", result.AgentStatus, result.ErrorCode, result.ErrorMessage)
	}
	if !result.RepairApplied {
		t.Error("expected RepairApplied=true")
	}
	if !result.RepairSuccess {
		t.Error("expected RepairSuccess=true")
	}
	if fc.calls != 2 {
		t.Errorf("expected 2 LLM calls, got %d", fc.calls)
	}
}

func TestPipeline_ParseFailed_RepairFailed(t *testing.T) {
	// Both calls return broken JSON.
	brokenJSON := `{"match_score": broken`
	fc := &fakeClient{responses: []string{brokenJSON, brokenJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusFallbackRequired {
		t.Fatalf("expected fallback_required, got: %s", result.AgentStatus)
	}
	if result.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got %s", result.GeneratedBy)
	}
	if !result.RepairApplied {
		t.Error("expected RepairApplied=true")
	}
	if result.RepairSuccess {
		t.Error("expected RepairSuccess=false")
	}
	if !result.FallbackRequired {
		t.Error("expected FallbackRequired=true")
	}
	if result.FallbackReason == "" {
		t.Error("expected non-empty FallbackReason")
	}
	if result.ErrorCode == "" {
		t.Error("expected non-empty ErrorCode")
	}
	if fc.calls != 2 {
		t.Errorf("expected 2 LLM calls (initial + repair), got %d", fc.calls)
	}
}

func TestPipeline_LLMError_FallbackRequired(t *testing.T) {
	fc := &fakeClient{err: fmt.Errorf("LLM service unavailable")}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusFallbackRequired {
		t.Fatalf("expected fallback_required, got: %s", result.AgentStatus)
	}
	if result.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got %s", result.GeneratedBy)
	}
	if result.FallbackReason != FallbackLLMCallFailed {
		t.Errorf("expected fallback_reason=%s, got %s", FallbackLLMCallFailed, result.FallbackReason)
	}
	if fc.calls != 1 {
		t.Errorf("expected 1 LLM call, got %d", fc.calls)
	}
}

func TestPipeline_LLMStructuredError_RateLimit(t *testing.T) {
	llmErr := llm.NewLLMError(llm.ErrLLMRateLimit, "rate limited")
	fc := &fakeClient{err: llmErr}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusFallbackRequired {
		t.Fatalf("expected fallback_required, got: %s", result.AgentStatus)
	}
	if result.ErrorCode != ErrLLMRateLimit {
		t.Errorf("expected error_code=%s, got %s", ErrLLMRateLimit, result.ErrorCode)
	}
}

func TestPipeline_UnknownTargetOutputType(t *testing.T) {
	fc := &fakeClient{}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "TestAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetOutputType("unknown"),
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusFallbackRequired {
		t.Fatalf("expected fallback_required, got: %s", result.AgentStatus)
	}
	if result.ErrorCode != ErrPipelineRequestInvalid {
		t.Errorf("expected error_code=%s, got %s", ErrPipelineRequestInvalid, result.ErrorCode)
	}
	if fc.calls != 0 {
		t.Errorf("expected 0 LLM calls, got %d", fc.calls)
	}
}

func TestPipeline_PromptLoadFailed(t *testing.T) {
	fc := &fakeClient{}
	promptDir, _ := filepath.Abs("/nonexistent/prompts")
	schemaDir, _ := filepath.Abs("../../../prompts/schema")

	pl := prompt.NewLoader(promptDir)
	sl := schema.NewLoader(schemaDir)
	v := schema.NewValidator()
	repairAgent := repair.NewAgent(fc, pl, sl, v)
	runner := NewRunner(fc, pl, sl, v, repairAgent)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "TestAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusFallbackRequired {
		t.Fatalf("expected fallback_required, got: %s", result.AgentStatus)
	}
	if result.ErrorCode != ErrPromptLoadFailed {
		t.Errorf("expected error_code=%s, got %s", ErrPromptLoadFailed, result.ErrorCode)
	}
	if fc.calls != 0 {
		t.Errorf("expected 0 LLM calls, got %d", fc.calls)
	}
}

func TestPipeline_SchemaLoadFailed(t *testing.T) {
	fc := &fakeClient{responses: []string{"{}"}}
	promptDir, _ := filepath.Abs("../../../prompts")
	schemaDir, _ := filepath.Abs("/nonexistent/schema")

	pl := prompt.NewLoader(promptDir)
	sl := schema.NewLoader(schemaDir)
	v := schema.NewValidator()
	repairAgent := repair.NewAgent(fc, pl, sl, v)
	runner := NewRunner(fc, pl, sl, v, repairAgent)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "TestAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusFallbackRequired {
		t.Fatalf("expected fallback_required, got: %s", result.AgentStatus)
	}
	if result.ErrorCode != ErrSchemaLoadFailed {
		t.Errorf("expected error_code=%s, got %s", ErrSchemaLoadFailed, result.ErrorCode)
	}
}

func TestPipeline_RepairOnce_MaxTwoCalls(t *testing.T) {
	// Both calls return broken JSON; repair is attempted exactly once.
	brokenJSON := `not json {{{`
	fc := &fakeClient{responses: []string{brokenJSON, brokenJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusFallbackRequired {
		t.Fatalf("expected fallback_required, got: %s", result.AgentStatus)
	}
	// Initial call + 1 repair call = 2 total.
	if fc.calls != 2 {
		t.Errorf("repair once constraint: expected 2 LLM calls (initial + repair), got %d", fc.calls)
	}
	if !result.RepairApplied {
		t.Error("expected RepairApplied=true")
	}
}

func TestPipeline_FallbackRequired_DoesNotCallMock(t *testing.T) {
	// When fallback_required, the pipeline only returns the decision.
	// It does NOT call any Mock Agent.
	fc := &fakeClient{err: fmt.Errorf("LLM unavailable")}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if !result.FallbackRequired {
		t.Error("expected FallbackRequired=true")
	}
	// Pipeline should not produce a Mock result — only the decision.
	if result.Parsed != nil {
		t.Error("fallback_required should not have Parsed data")
	}
	if result.JSONText != "" {
		t.Error("fallback_required should not have JSONText")
	}
}

func TestPipeline_ErrorDoesNotContainFakeAPIKey(t *testing.T) {
	fc := &fakeClient{err: fmt.Errorf("auth failed with key=sk-fake-1234567890")}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "TestAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if strings.Contains(result.ErrorMessage, "sk-fake-1234567890") {
		t.Error("error message must not contain fake API key")
	}
}

func TestPipeline_RawOutputLength(t *testing.T) {
	validJSON := toJSON(minimalMatchAnalysis())
	fc := &fakeClient{responses: []string{validJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if !result.FallbackRequired && result.RawOutputLength <= 0 {
		t.Error("expected RawOutputLength > 0 on success")
	}
}

func TestPipeline_RepairAppliedAndRepairSuccess(t *testing.T) {
	brokenJSON := `{"match_score": broken`
	validJSON := toJSON(minimalMatchAnalysis())
	fc := &fakeClient{responses: []string{brokenJSON, validJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if !result.RepairApplied {
		t.Error("expected RepairApplied=true")
	}
	if !result.RepairSuccess {
		t.Error("expected RepairSuccess=true")
	}
	if result.AgentStatus != StatusRepaired {
		t.Errorf("expected AgentStatus=repaired, got %s", result.AgentStatus)
	}
}

func TestPipeline_LLMDisabled(t *testing.T) {
	llmErr := llm.NewLLMError(llm.ErrLLMDisabled, "LLM is disabled")
	fc := &fakeClient{err: llmErr}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "TestAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.AgentStatus != StatusFallbackRequired {
		t.Fatalf("expected fallback_required, got: %s", result.AgentStatus)
	}
	if result.ErrorCode != ErrLLMDisabled {
		t.Errorf("expected error_code=%s, got %s", ErrLLMDisabled, result.ErrorCode)
	}
	if result.FallbackReason != FallbackLLMDisabled {
		t.Errorf("expected fallback_reason=%s, got %s", FallbackLLMDisabled, result.FallbackReason)
	}
}

func TestPipeline_DurationMs(t *testing.T) {
	validJSON := toJSON(minimalMatchAnalysis())
	fc := &fakeClient{responses: []string{validJSON}}
	runner := newTestRunner(fc)

	result := runner.Run(context.Background(), RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: TargetMatchAnalysis,
		PromptValues:     map[string]string{},
	})

	if result.DurationMs <= 0 {
		t.Errorf("expected DurationMs > 0, got %d", result.DurationMs)
	}
}

// Ensure unused import is not flagged
var _ = time.Millisecond
