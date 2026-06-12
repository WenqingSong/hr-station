package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// ─── SEP Profiler ────────────────────────────────────────────────────────

// SemanticProfileProfiler handles Hy3-based semantic profile generation for resumes and jobs.
type SemanticProfileProfiler struct {
	semanticService *SemanticProfileService
	resumeRepo      *repository.ResumeRepository
	jobRepo         *repository.JobRepository
	llmClient       llm.Client
	qualityModel    string
}

// NewSemanticProfileProfiler creates a new profiler.
func NewSemanticProfileProfiler(semanticService *SemanticProfileService, resumeRepo *repository.ResumeRepository, jobRepo *repository.JobRepository, llmClient llm.Client, qualityModel string) *SemanticProfileProfiler {
	return &SemanticProfileProfiler{
		semanticService: semanticService,
		resumeRepo:      resumeRepo,
		jobRepo:         jobRepo,
		llmClient:       llmClient,
		qualityModel:    qualityModel,
	}
}

// ─── Ensure Profile (reuse if verified, generate if not) ──────────────────

// EnsureResumeSemanticProfile returns a verified profile or generates one.
func (p *SemanticProfileProfiler) EnsureResumeSemanticProfile(ctx context.Context, userID uint64, resumeID int64) (*model.SemanticProfile, []*model.SemanticProfileTag, error) {
	hash, err := p.semanticService.ComputeResumeContentHash(ctx, resumeID)
	if err != nil {
		return nil, nil, fmt.Errorf("ensure resume profile: hash: %w", err)
	}

	// Try existing verified profile
	existing, tags, err := p.semanticService.GetVerifiedProfile(ctx, model.SemanticTargetResume, resumeID, hash)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return existing, tags, nil
	}

	return p.GenerateResumeSemanticProfile(ctx, userID, resumeID)
}

// GenerateResumeSemanticProfile generates a fresh resume semantic profile via Hy3 high.
func (p *SemanticProfileProfiler) GenerateResumeSemanticProfile(ctx context.Context, userID uint64, resumeID int64) (*model.SemanticProfile, []*model.SemanticProfileTag, error) {
	// Fetch resume data
	var resume *model.Resume
	if userID > 0 {
		resumes, err := p.resumeRepo.ListByUserID(ctx, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("generate resume profile: fetch: %w", err)
		}
		for _, r := range resumes {
			if r.ID == uint64(resumeID) {
				resume = r
				break
			}
		}
	}
	if resume == nil {
		// Fallback: try ListAllIDPairs to verify existence, then read directly
		pairs, err := p.resumeRepo.ListAllIDPairs(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("generate resume profile: list all: %w", err)
		}
		for _, pair := range pairs {
			if pair.ID == uint64(resumeID) {
				userID = pair.UserID
				break
			}
		}
		if userID == 0 {
			return nil, nil, fmt.Errorf("resume %d not found", resumeID)
		}
		resumes, err := p.resumeRepo.ListByUserID(ctx, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("generate resume profile: fetch by user: %w", err)
		}
		for _, r := range resumes {
			if r.ID == uint64(resumeID) {
				resume = r
				break
			}
		}
	}
	if resume == nil {
		return nil, nil, fmt.Errorf("resume %d not found", resumeID)
	}

	// Compute hash
	hash, err := p.semanticService.ComputeResumeContentHash(ctx, resumeID)
	if err != nil {
		return nil, nil, fmt.Errorf("generate resume profile: hash: %w", err)
	}

	// Mark stale existing
	_ = p.semanticService.MarkStale(ctx, model.SemanticTargetResume, resumeID)

	// Build prompts
	sourceText := BuildSourceFromResume(resume.Name, resume.RawText, resume.UpdatedAt)
	sysPrompt, usrPrompt := BuildResumeSemanticProfilePrompt(resumeID, resume.Name, sourceText)

	// Call Hy3
	output, err := p.callLLMForProfile(ctx, model.SemanticTargetResume, resumeID, hash, sysPrompt, usrPrompt)
	if err != nil {
		p.saveFailedProfile(ctx, model.SemanticTargetResume, resumeID, hash, err.Error())
		return nil, nil, err
	}

	return p.persistProfile(ctx, model.SemanticTargetResume, resumeID, hash, output, model.SemanticPromptProfileResumeV1)
}

// ─── Job Profiler ──────────────────────────────────────────────────────────

// EnsureJobSemanticProfile returns a verified job profile or generates one.
func (p *SemanticProfileProfiler) EnsureJobSemanticProfile(ctx context.Context, jobID int64) (*model.SemanticProfile, []*model.SemanticProfileTag, error) {
	hash, err := p.semanticService.ComputeJobContentHash(ctx, jobID)
	if err != nil {
		return nil, nil, fmt.Errorf("ensure job profile: hash: %w", err)
	}

	existing, tags, err := p.semanticService.GetVerifiedProfile(ctx, model.SemanticTargetJob, jobID, hash)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return existing, tags, nil
	}

	return p.GenerateJobSemanticProfile(ctx, jobID)
}

// GenerateJobSemanticProfile generates a fresh job semantic profile via Hy3 high.
func (p *SemanticProfileProfiler) GenerateJobSemanticProfile(ctx context.Context, jobID int64) (*model.SemanticProfile, []*model.SemanticProfileTag, error) {
	job, err := p.jobRepo.FindByID(ctx, uint64(jobID))
	if err != nil {
		return nil, nil, fmt.Errorf("generate job profile: fetch: %w", err)
	}
	if job == nil {
		return nil, nil, fmt.Errorf("job %d not found", jobID)
	}

	hash, err := p.semanticService.ComputeJobContentHash(ctx, jobID)
	if err != nil {
		return nil, nil, fmt.Errorf("generate job profile: hash: %w", err)
	}

	_ = p.semanticService.MarkStale(ctx, model.SemanticTargetJob, jobID)

	sysPrompt, usrPrompt := BuildJobSemanticProfilePrompt(jobID, job.Title, job.Description, job.Requirements, job.Direction, job.Category)

	output, err := p.callLLMForProfile(ctx, model.SemanticTargetJob, jobID, hash, sysPrompt, usrPrompt)
	if err != nil {
		p.saveFailedProfile(ctx, model.SemanticTargetJob, jobID, hash, err.Error())
		return nil, nil, err
	}

	return p.persistProfile(ctx, model.SemanticTargetJob, jobID, hash, output, model.SemanticPromptProfileJobV1)
}

// ─── Internal ───────────────────────────────────────────────────────────────

func (p *SemanticProfileProfiler) callLLMForProfile(ctx context.Context, targetType string, targetID int64, hash, sysPrompt, usrPrompt string) (*model.SemanticProfileLLMOutput, error) {
	req := llm.ChatRequest{
		Model:           p.qualityModel,
		Messages:        []llm.Message{{Role: llm.RoleSystem, Content: sysPrompt}, {Role: llm.RoleUser, Content: usrPrompt}},
		Temperature:     0.3,
		MaxTokens:       4096,
		ReasoningEffort: "high",
	}
	start := time.Now()
	resp, err := p.llmClient.Chat(ctx, req)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return nil, fmt.Errorf("hy3 profiler: %s %d: chat failed after %dms: %w", targetType, targetID, elapsed, err)
	}
	log.Printf("[SEP-Profiler] %s %d: hy3 high completed in %dms, tokens=%d", targetType, targetID, elapsed, resp.Usage.TotalTokens)

	// Extract JSON from response
	content := extractJSONFromContent(resp.RawContent)

	var output model.SemanticProfileLLMOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		// Attempt repair with simpler extraction
		content = strings.TrimSpace(content)
		if !strings.HasPrefix(content, "{") {
			if idx := strings.Index(content, "{"); idx >= 0 {
				content = content[idx:]
			}
		}
		if !strings.HasSuffix(content, "}") {
			if idx := strings.LastIndex(content, "}"); idx >= 0 {
				content = content[:idx+1]
			}
		}
		if err2 := json.Unmarshal([]byte(content), &output); err2 != nil {
			return nil, fmt.Errorf("hy3 profiler: %s %d: JSON parse failed after repair: %w (raw: %s...)", targetType, targetID, err, truncate(resp.RawContent, 200))
		}
	}

	// Sanitize
	output.SanitizeLLMOutput()

	// Validate
	if err := ValidateSemanticProfileOutput(&output, targetType); err != nil {
		return nil, fmt.Errorf("hy3 profiler: %s %d: validation failed: %w", targetType, targetID, err)
	}

	return &output, nil
}

func (p *SemanticProfileProfiler) persistProfile(ctx context.Context, targetType string, targetID int64, hash string, output *model.SemanticProfileLLMOutput, promptVersion string) (*model.SemanticProfile, []*model.SemanticProfileTag, error) {
	now := time.Now()
	profile := &model.SemanticProfile{
		TargetType:      targetType,
		TargetID:        targetID,
		ProfileVersion:  model.SemanticProfileVersionV1,
		PromptVersion:   promptVersion,
		SchemaVersion:   model.SemanticSchemaVersionV1,
		ContentHash:     hash,
		ModelProvider:   "hy3",
		ModelName:       p.qualityModel,
		ReasoningEffort: "high",
		Status:          model.SemanticProfileStatusVerified,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	output.ToPersistenceJSON(profile)

	tags := make([]*model.SemanticProfileTag, 0, len(output.Tags))
	for _, t := range output.Tags {
		tags = append(tags, &model.SemanticProfileTag{
			ProfileID:     0, // set by repository after insert
			TargetType:    targetType,
			TargetID:      targetID,
			TagCode:       t.TagCode,
			TagLabel:      t.TagLabel,
			Category:      t.Category,
			Confidence:    t.Confidence,
			EvidenceText:  t.EvidenceText,
			Explanation:   t.Explanation,
			SourceSection: t.SourceSection,
			IsUserVisible: t.IsUserVisible,
			Weight:        t.Weight,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}

	if err := p.semanticService.SaveVerifiedProfile(ctx, profile, tags); err != nil {
		return nil, nil, fmt.Errorf("persist %s profile: %w", targetType, err)
	}

	log.Printf("[SEP-Profiler] %s %d: verified profile id=%d tags=%d summary=%.80s", targetType, targetID, profile.ID, len(tags), profile.SummarySentence)
	return profile, tags, nil
}

func (p *SemanticProfileProfiler) saveFailedProfile(ctx context.Context, targetType string, targetID int64, hash, errMsg string) {
	now := time.Now()
	profile := &model.SemanticProfile{
		TargetType:      targetType,
		TargetID:        targetID,
		ProfileVersion:  model.SemanticProfileVersionV1,
		PromptVersion:   model.SemanticPromptProfileGenericV1,
		SchemaVersion:   model.SemanticSchemaVersionV1,
		ContentHash:     hash,
		ModelProvider:   "hy3",
		ModelName:       p.qualityModel,
		ReasoningEffort: "high",
		Status:          model.SemanticProfileStatusFailed,
		ErrorMessage:    errMsg,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_ = p.semanticService.semanticRepo.CreateProfile(ctx, profile)
	log.Printf("[SEP-Profiler] %s %d: failed profile saved, error=%s", targetType, targetID, errMsg)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func extractJSONFromContent(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```json") {
		raw = strings.TrimPrefix(raw, "```json")
		if idx := strings.LastIndex(raw, "```"); idx >= 0 {
			raw = raw[:idx]
		}
	} else if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```")
		if idx := strings.LastIndex(raw, "```"); idx >= 0 {
			raw = raw[:idx]
		}
	}
	return strings.TrimSpace(raw)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen { return s }
	return s[:maxLen] + "..."
}
