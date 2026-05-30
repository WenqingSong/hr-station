package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// mapping maps each SchemaName to its file name relative to baseDir.
var mapping = map[SchemaName]string{
	SchemaMatchAnalysis:      "match_analysis_schema.json",
	SchemaRecommendation:     "recommendation_schema.json",
	SchemaResumeOptimization: "resume_optimization_schema.json",
}

// Loader loads JSON Schema files from a configured directory.
type Loader struct {
	baseDir string
}

// NewLoader creates a new Schema Loader.
// baseDir should point to the directory containing schema .json files
// (e.g. "server/prompts/schema").
func NewLoader(baseDir string) *Loader {
	return &Loader{baseDir: baseDir}
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
