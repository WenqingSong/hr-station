package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// resolveDir tries to find an existing directory matching baseDir.
// It tries the baseDir as-is first, then strips a "server/" prefix
// to handle the case where the process runs from the server/ directory
// but the config uses "server/prompts/schema" (meant for project root).
func resolveDir(baseDir string) string {
	// Try the baseDir as-is first.
	if fi, err := os.Stat(baseDir); err == nil && fi.IsDir() {
		return baseDir
	}

	// If baseDir starts with "server/" or "server\", try without the prefix.
	stripped := ""
	if strings.HasPrefix(baseDir, "server/") {
		stripped = baseDir[7:]
	} else if strings.HasPrefix(baseDir, "server\\") {
		stripped = baseDir[7:]
	}
	if stripped != "" {
		if fi, err := os.Stat(stripped); err == nil && fi.IsDir() {
			return stripped
		}
	}

	// Return original — will fail later with a clear file-not-found error.
	return baseDir
}

// mapping maps each SchemaName to its file name relative to baseDir.
var mapping = map[SchemaName]string{
	SchemaMatchAnalysis:      "match_analysis_schema.json",
	SchemaRecommendation:     "recommendation_schema.json",
	SchemaResumeOptimization: "resume_optimization_schema.json",
	SchemaVersionCompare:     "version_compare_schema.json",
	SchemaResumeOptimizationCore: "resume_optimization_core_schema.json",
	// K13-Fix-C1: Deep recommendation 3-agent schemas
	SchemaRecommendationDecision: "recommendation_decision_schema.json",
	SchemaRecommendationVerifier: "recommendation_verifier_schema.json",
	// K13-Fix-C2: Deep match 3-agent schemas
	SchemaMatchDecision: "match_decision_schema.json",
	SchemaMatchVerifier: "match_verifier_schema.json",
	// K13-Fix-C3: Deep optimization 3-agent schemas
	SchemaOptimizationDecision: "optimization_decision_schema.json",
	SchemaOptimizationVerifier: "optimization_verifier_schema.json",
	// K13-Fix-C4: Deep compare 3-agent schemas
	SchemaCompareDecision: "compare_decision_schema.json",
	SchemaCompareVerifier: "compare_verifier_schema.json",
}

// Loader loads JSON Schema files from a configured directory.
type Loader struct {
	baseDir string
}

// NewLoader creates a new Schema Loader.
// baseDir should point to the directory containing schema .json files
// (e.g. "server/prompts/schema").
// If baseDir does not exist as a directory, NewLoader tries alternative
// paths (e.g. stripping a "server/" prefix) before giving up.
func NewLoader(baseDir string) *Loader {
	return &Loader{baseDir: resolveDir(baseDir)}
}

// Load reads and parses the JSON Schema for the given logical name.
// It validates that the file contains a valid JSON object (top-level {}).
// It does NOT perform full JSON Schema validation — that is 4-2.3.
func (l *Loader) Load(name SchemaName) (*Document, error) {
	fileName, ok := mapping[name]
	if !ok {
		return nil, fmt.Errorf("schema: unknown schema name %q", name)
	}

	path := l.resolvePath(fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("schema: failed to read %s: %w", path, err)
	}

	// Parse as JSON and verify top-level is an object.
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("schema: invalid JSON in %s: %w", path, err)
	}

	return &Document{
		Name:    name,
		Path:    path,
		Content: string(data),
		Raw:     raw,
	}, nil
}

// List returns all known SchemaNames in sorted order.
func (l *Loader) List() []SchemaName {
	names := make([]SchemaName, 0, len(mapping))
	for n := range mapping {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool {
		return string(names[i]) < string(names[j])
	})
	return names
}

// ResolvePath returns the full file path for the given logical name.
// Returns an error if the name is unknown.
func (l *Loader) ResolvePath(name SchemaName) (string, error) {
	fileName, ok := mapping[name]
	if !ok {
		return "", fmt.Errorf("schema: unknown schema name %q", name)
	}
	return l.resolvePath(fileName), nil
}

// Title returns the "title" field from the Schema, or "" if absent.
func (d *Document) Title() string {
	if v, ok := d.Raw["title"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// RequiredCount returns the number of fields in the "required" array,
// or 0 if absent.
func (d *Document) RequiredCount() int {
	if v, ok := d.Raw["required"]; ok {
		if arr, ok := v.([]any); ok {
			return len(arr)
		}
	}
	return 0
}

// Draft returns the "$schema" URI, or "" if absent.
func (d *Document) Draft() string {
	if v, ok := d.Raw["$schema"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// resolvePath joins baseDir with fileName.
// It also verifies the resulting path stays within baseDir (no ../ traversal).
func (l *Loader) resolvePath(fileName string) string {
	joined := filepath.Join(l.baseDir, fileName)

	absBase, _ := filepath.Abs(l.baseDir)
	absJoined, _ := filepath.Abs(joined)
	if strings.HasPrefix(absJoined, absBase+string(filepath.Separator)) || absJoined == absBase {
		return joined
	}
	return joined
}
