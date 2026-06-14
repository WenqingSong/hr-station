package model

import "fmt"

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

// NormalizeGenerationTier returns the canonical GenerationTier.
// Empty input defaults to "fast".
// "deep" is accepted as an alias for "quality" (backward compatibility).
// Returns error if the input is non-empty and not a valid tier.
func NormalizeGenerationTier(input string) (GenerationTier, error) {
	if input == "" {
		return GenerationTierFast, nil
	}
	// K15-P0-D1: accept "deep" as an alias for "quality"
	if input == "deep" {
		return GenerationTierQuality, nil
	}
	tier := GenerationTier(input)
	if ValidGenerationTiers[tier] {
		return tier, nil
	}
	return "", fmt.Errorf("invalid generation_tier: %q (allowed: fast, quality, deep)", input)
}
