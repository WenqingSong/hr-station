// ─── H19-B: Deep Match Agent ──────────────────────────────────────────────
// Wraps existing 3-agent pipeline with H19-B post-processing:
//   - fix null arrays, sanitize risks, ensure generated_by=deep_match_agent
//   - store DeepMatchReport in result_json for future H19-C UI

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
)

// ─── Main Entry ──────────────────────────────────────────────────────────

func (s *MatchService) runDeepMatchSEP(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job, runner agent.PipelineRunner,
) (*MatchResponse, error) {
	log.Printf("[DeepMatchSEP] starting for resume=%d job=%d", resume.ID, job.ID)

	var resp *MatchResponse
	var err error

	// Try 3-agent via tag evidence (existing pipeline)
	if s.tagEvidence != nil {
		resp, err = s.matchWith3Agents(ctx, userID, resume, job, runner)
		if err == nil {
			// H19-B: post-process to meet new contract
			resp = sanitizeDeepResponse(resp)
			return resp, nil
		}
		log.Printf("[DeepMatchSEP] 3-agent failed: %v", err)
	}

	// Fallback to single-agent LLM
	resp, fallbackErr := s.analyzeWithLLM(ctx, resume, job, runner)
	if fallbackErr == nil {
		resp = sanitizeDeepResponse(resp)
		return resp, nil
	}
	log.Printf("[DeepMatchSEP] LLM fallback failed: %v", fallbackErr)

	// Last resort
	return s.fallbackDeepMatch(ctx, userID, resume, job)
}

// ─── Post-process: enforce H19-B contract ────────────────────────────────

func sanitizeDeepResponse(resp *MatchResponse) *MatchResponse {
	if resp == nil { return resp }
	resp.GeneratedBy = "deep_match_agent"
	if resp.Result == nil { return resp }

	r := resp.Result
	r.GeneratedBy = "deep_match_agent"
	// H19-D: also normalize "repaired" → "match_verified"
	if r.AgentStatus == "" || r.AgentStatus == "success" || r.AgentStatus == "repaired" {
		r.AgentStatus = "match_verified"
	}
	// Fix null arrays
	if r.MatchHighlights == nil { r.MatchHighlights = []string{} }
	if r.GapPoints == nil { r.GapPoints = []string{} }
	if r.RiskPoints == nil { r.RiskPoints = []string{} }
	if r.EvidenceChain == nil { r.EvidenceChain = []model.MatchEvidence{} }
	if r.NextSteps == nil { r.NextSteps = []string{} }
	if r.OptimizationSuggestions == nil { r.OptimizationSuggestions = []model.MatchSuggestion{} }
	if r.FitPoints == nil { r.FitPoints = []string{} }
	if r.JobRequirements == nil { r.JobRequirements = []string{} }
	if r.ResumeEvidence == nil { r.ResumeEvidence = []string{} }

	// Sanitize risk strings: replace bare tag-like strings with human labels
	cleanRisks := []string{}
	for _, risk := range r.RiskPoints {
		cleaned := sanitizeRiskText(risk)
		cleanRisks = append(cleanRisks, cleaned)
	}
	r.RiskPoints = cleanRisks

	// Fix level
	if r.MatchLevel == "" || r.MatchLevel == "low" {
		r.MatchLevel = matchLevel(r.MatchScore)
	}

	// Fix summary
	if r.Summary == "" || r.Summary == r.MatchLevel {
		r.Summary = fmt.Sprintf("深度分析匹配度%d分（%s）", r.MatchScore, r.MatchLevel)
	}

	// Update meta
	resp.Meta.GeneratedBy = "deep_match_agent"
	resp.Meta.AgentStatus = r.AgentStatus

	return resp
}

func sanitizeRiskText(s string) string {
	// Map legacy risk strings to better labels
	mappings := map[string]string{
		"学历风险": "学历要求可能不满足岗位筛选条件",
		"核心技术缺失风险": "核心技能覆盖不足，可能影响简历筛选通过率",
		"论文要求风险": "岗位可能要求论文/研究成果，简历未体现",
		"实习风险": "缺少相关实习经历，竞争力可能不足",
		"方向风险": "岗位方向与简历核心方向存在差异",
	}
	if v, ok := mappings[s]; ok { return v }
	return s
}

// ─── Fallback ─────────────────────────────────────────────────────────────

func (s *MatchService) fallbackDeepMatch(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job,
) (*MatchResponse, error) {
	log.Printf("[DeepMatchSEP] fallback for resume=%d job=%d", resume.ID, job.ID)
	var legacyResult *model.MatchAnalysisResult
	if s.semanticRepo != nil {
		rp, _ := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(resume.ID))
		jp, _ := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(job.ID))
		if rp != nil && jp != nil && rp.Status == "verified" && jp.Status == "verified" {
			rtags, _ := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
			jtgs, _ := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
			legacyResult = runFastSEPMatch(&MatchSEPContext{
				ResumeTags: rtags, JobTags: jtgs, JobTitle: job.Title,
				ResumeID: int64(resume.ID), JobID: int64(job.ID),
			})
		}
	}
	if legacyResult == nil {
		legacyResult = &model.MatchAnalysisResult{
			MatchScore: 50, MatchLevel: "中等匹配", GeneratedBy: "deep_match_agent", AgentStatus: "match_repaired",
			Summary: "深度分析暂时不可用，当前为安全兜底结果。",
			MatchHighlights: []string{}, GapPoints: []string{}, RiskPoints: []string{},
			NextSteps: []string{"建议稍后重试深度分析"},
		}
	} else {
		legacyResult.GeneratedBy = "deep_match_agent"
		legacyResult.AgentStatus = "match_repaired"
	}
	meta := model.AgentResponseMeta{
		AgentName: "DeepMatchAgent", GeneratedBy: "deep_match_agent", AgentStatus: "match_repaired",
		FallbackReason: "llm_unavailable", RepairApplied: true, SchemaValid: false,
	}
	_ = json.Marshal // keep import
	return &MatchResponse{
		AnalysisID: 0, GeneratedBy: "deep_match_agent", Result: legacyResult,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"), Meta: meta,
	}, nil
}


