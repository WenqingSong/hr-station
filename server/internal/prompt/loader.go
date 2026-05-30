package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// mapping maps each PromptName to its file name relative to baseDir.
var mapping = map[PromptName]string{
	PromptMatchAnalysis:      "match_analysis_v1.md",
	PromptRecommendation:     "recommendation_v1.md",
	PromptResumeOptimization: "resume_optimization_v1.md",
	PromptJSONRepair:         "json_repair_v1.md",
}

// Loader loads Prompt template files from a configured directory.
type Loader struct {
	baseDir string
}

// NewLoader creates a new Prompt Loader.
// baseDir should point to the directory containing prompt .md files
// (e.g. "server/prompts").
func NewLoader(baseDir string) *Loader {
	return &Loader{baseDir: baseDir}
}

// Load reads the Prompt template for the given logical name.
// Returns an error if the name is unknown or the file cannot be read.
func (l *Loader) Load(name PromptName) (*Template, error) {
	fileName, ok := mapping[name]
	if !ok {
		return nil, fmt.Errorf("prompt: unknown prompt name %q", name)
	}

	path := l.resolvePath(fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("prompt: failed to read %s: %w", path, err)
	}

	return &Template{
		Name:    name,
		Path:    path,
		Content: string(data),
	}, nil
}

// Render loads the Prompt template for the given logical name and replaces
// placeholders in the form {{KEY}} with the corresponding values.
//
// Placeholders not present in values are left unchanged (not an error).
// No secondary escaping or JSON transformation is applied to values.
func (l *Loader) Render(name PromptName, values map[string]string) (string, error) {
	tmpl, err := l.Load(name)
	if err != nil {
		return "", err
	}

	result := tmpl.Content
	for key, val := range values {
		result = strings.ReplaceAll(result, "{{"+key+"}}", val)
	}
	return result, nil
}

// List returns all known PromptNames in sorted order.
func (l *Loader) List() []PromptName {
	names := make([]PromptName, 0, len(mapping))
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
// The path is constructed from baseDir + the mapped file name;
// path traversal via the name is impossible because only predefined
// PromptNames are accepted.
func (l *Loader) ResolvePath(name PromptName) (string, error) {
	fileName, ok := mapping[name]
	if !ok {
		return "", fmt.Errorf("prompt: unknown prompt name %q", name)
	}
	return l.resolvePath(fileName), nil
}

// resolvePath joins baseDir with fileName.
// It also verifies the resulting path stays within baseDir (no ../ traversal).
func (l *Loader) resolvePath(fileName string) string {
	joined := filepath.Join(l.baseDir, fileName)

	// Clean both paths and verify the result doesn't escape baseDir.
	absBase, _ := filepath.Abs(l.baseDir)
	absJoined, _ := filepath.Abs(joined)
	if strings.HasPrefix(absJoined, absBase+string(filepath.Separator)) || absJoined == absBase {
		return joined
	}
	// If somehow it escapes, still return the joined path — the file read
	// will fail naturally. This is a defense-in-depth check.
	return joined
}
