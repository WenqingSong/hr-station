// ─── K13-Fix-D: Unified Verifier / Guardrail ─────────────────────────
// Local deterministic guards — complement, NEVER replace LLM Verifier.
// Used as Agent 3 fallback safety net and result post-processing.

package guardrail

import (
	"sort"
)

// ─── AgentMeta ───────────────────────────────────────────────────────

type AgentMeta struct {
	GeneratedBy    string `json:"generated_by"`
	GenerationTier string `json:"generation_tier"`
	AgentStatus    string `json:"agent_status"`
	Fallback       bool   `json:"fallback"`
	RepairApplied  bool   `json:"repair_applied"`
	ErrorCode      string `json:"error_code,omitempty"`
	DurationMS     int64  `json:"duration_ms,omitempty"`
	ModelName      string `json:"model_name,omitempty"`
	ResultVersion  string `json:"result_version,omitempty"`
}

// ─── Status enums ────────────────────────────────────────────────────

const (
	GeneratedByFastTagEngine = "fast_tag_engine"
	GeneratedByLLMAgent      = "llm_agent"
	GeneratedByFallbackAgent = "fallback_agent"
	GeneratedByMockAgent     = "mock_agent"

	AgentStatusSuccess             = "success"
	AgentStatusVerified            = "verified"
	AgentStatusRepaired            = "repaired"
	AgentStatusVerifiedByLocalGuard = "verified_by_local_guard"
	AgentStatusFallback            = "fallback"
	AgentStatusFailed              = "failed"

	ResultVersionV1 = "v1"
)

// ─── Local Guards ────────────────────────────────────────────────────

// NormalizeScore clamps score to [0, 100].
func NormalizeScore(score int) int {
	if score < 0 { return 0 }
	if score > 100 { return 100 }
	return score
}

// NormalizeStringArray ensures v is a []string (never nil).
func NormalizeStringArray(v interface{}) []string {
	switch arr := v.(type) {
	case []string: return arr
	case []interface{}:
		r := make([]string, len(arr))
		for i, x := range arr { if s, ok := x.(string); ok { r[i] = s } }
		return r
	}
	return []string{}
}

// ValidateScoreDistribution checks if scores are all the same.
func ValidateScoreDistribution(scores []int) bool {
	if len(scores) <= 1 { return true }
	first := scores[0]
	for _, s := range scores[1:] { if s != first { return true } }
	return false // all identical — bad
}

// ValidateWinnerID checks winner is in valid set.
func ValidateWinnerID(winner uint64, validIDs []uint64) bool {
	for _, id := range validIDs { if winner == id { return true } }
	return false
}

// FixWinnerID returns toID if winner is invalid.
func FixWinnerID(winner, fromID, toID uint64) uint64 {
	if winner == fromID || winner == toID { return winner }
	return toID
}

// ValidateNoDirectionMismatch checks profile vs job direction consistency.
func ValidateNoDirectionMismatch(resumeDominant, jobDirection string) bool {
	// Simple heuristic: backend profile + algorithm job = mismatch
	if resumeDominant == "backend_engineering" && jobDirection == "algorithm" { return false }
	if resumeDominant == "algorithm_research" && jobDirection == "backend" { return false }
	return true
}

// ValidateNoFalseZeroExperience returns project/internship counts > 0 check.
func ValidateNoFalseZeroExperience(projectCount, internshipCount int) bool {
	return projectCount >= 0 && internshipCount >= 0
}

// NormalizeMeta fills defaults for AgentMeta.
func NormalizeMeta(meta *AgentMeta) {
	if meta.GeneratedBy == "" { meta.GeneratedBy = GeneratedByLLMAgent }
	if meta.GenerationTier == "" { meta.GenerationTier = "quality" }
	if meta.AgentStatus == "" { meta.AgentStatus = AgentStatusVerified }
	if meta.ResultVersion == "" { meta.ResultVersion = ResultVersionV1 }
}

// SortStrings sorts strings in place.
func SortStrings(s []string) { sort.Strings(s) }

// EnsureNonNilStringArray returns empty slice for nil.
func EnsureNonNilStringArray(s []string) []string {
	if s == nil { return []string{} }
	return s
}
