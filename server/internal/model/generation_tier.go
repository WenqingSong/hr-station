package model

// GenerationTier controls which LLM model tier is used for generation.
// Each core feature supports both fast and quality results.
type GenerationTier string

const (
	// GenerationTierFast uses deepseek-v4-flash for quick, low-cost results.
	GenerationTierFast GenerationTier = "fast"

	// GenerationTierQuality uses deepseek-v4-pro-202606 for deep, high-quality results.
	GenerationTierQuality GenerationTier = "quality"
)

// ValidGenerationTiers contains all allowed GenerationTier values.
var ValidGenerationTiers = map[GenerationTier]bool{
	GenerationTierFast:    true,
	GenerationTierQuality: true,
}

// NormalizeGenerationTier returns the canonical GenerationTier, defaulting to "fast".
func NormalizeGenerationTier(input string) GenerationTier {
	tier := GenerationTier(input)
	if ValidGenerationTiers[tier] {
		return tier
	}
	return GenerationTierFast
}
