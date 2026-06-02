package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

// testBaseDir points to the real prompts directory for integration tests.
// Tests that need the actual prompt files use this; others create temp dirs.
var testBaseDir = func() string {
	// Resolve relative to the server directory.
	d, _ := filepath.Abs("../../prompts")
	return d
}()

func TestLoader_Load_MatchAnalysis(t *testing.T) {
	l := NewLoader(testBaseDir)
	tmpl, err := l.Load(PromptMatchAnalysis)
	if err != nil {
		t.Fatalf("Load(match_analysis) error: %v", err)
	}
	if tmpl.Name != PromptMatchAnalysis {
		t.Errorf("Name = %q, want %q", tmpl.Name, PromptMatchAnalysis)
	}
	if tmpl.Content == "" {
		t.Error("Content is empty")
	}
	if tmpl.Path == "" {
		t.Error("Path is empty")
	}
}

func TestLoader_Load_Recommendation(t *testing.T) {
	l := NewLoader(testBaseDir)
	tmpl, err := l.Load(PromptRecommendation)
	if err != nil {
		t.Fatalf("Load(recommendation) error: %v", err)
	}
	if tmpl.Content == "" {
		t.Error("Content is empty")
	}
}

func TestLoader_Load_ResumeOptimization(t *testing.T) {
	l := NewLoader(testBaseDir)
	tmpl, err := l.Load(PromptResumeOptimization)
	if err != nil {
		t.Fatalf("Load(resume_optimization) error: %v", err)
	}
	if tmpl.Content == "" {
		t.Error("Content is empty")
	}
}

func TestLoader_Load_JSONRepair(t *testing.T) {
	l := NewLoader(testBaseDir)
	tmpl, err := l.Load(PromptJSONRepair)
	if err != nil {
		t.Fatalf("Load(json_repair) error: %v", err)
	}
	if tmpl.Content == "" {
		t.Error("Content is empty")
	}
}

func TestLoader_Load_UnknownName(t *testing.T) {
	l := NewLoader(testBaseDir)
	_, err := l.Load(PromptName("nonexistent"))
	if err == nil {
		t.Fatal("expected error for unknown prompt name, got nil")
	}
}

func TestLoader_Render(t *testing.T) {
	// Create a temp prompt file with placeholders.
	dir := t.TempDir()
	content := "Hello {{NAME}}, your score is {{SCORE}}."
	err := os.WriteFile(filepath.Join(dir, "test_v1.md"), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Temporarily add a mapping for the test.
	testName := PromptName("test_prompt")
	mapping[testName] = "test_v1.md"
	defer delete(mapping, testName)

	l := NewLoader(dir)
	result, err := l.Render(testName, map[string]string{
		"NAME":  "Alice",
		"SCORE": "85",
	})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	expected := "Hello Alice, your score is 85."
	if result != expected {
		t.Errorf("Render = %q, want %q", result, expected)
	}
}

func TestLoader_Render_MissingPlaceholder(t *testing.T) {
	// Placeholders not in values should remain unchanged.
	dir := t.TempDir()
	content := "Hello {{NAME}}, {{UNKNOWN}}."
	err := os.WriteFile(filepath.Join(dir, "test_v1.md"), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	testName := PromptName("test_missing")
	mapping[testName] = "test_v1.md"
	defer delete(mapping, testName)

	l := NewLoader(dir)
	result, err := l.Render(testName, map[string]string{
		"NAME": "Bob",
	})
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	expected := "Hello Bob, {{UNKNOWN}}."
	if result != expected {
		t.Errorf("Render = %q, want %q", result, expected)
	}
}

func TestLoader_List(t *testing.T) {
	l := NewLoader(testBaseDir)
	names := l.List()

	if len(names) != 6 {
		t.Fatalf("List returned %d names, want 6", len(names))
	}

	expected := map[PromptName]bool{
		PromptMatchAnalysis:          true,
		PromptRecommendation:         true,
		PromptResumeOptimization:     true,
		PromptJSONRepair:             true,
		PromptVersionCompare:         true,
		PromptResumeOptimizationCore: true,
	}
	for _, n := range names {
		if !expected[n] {
			t.Errorf("unexpected name %q", n)
		}
	}

	// Verify sorted order.
	for i := 1; i < len(names); i++ {
		if string(names[i-1]) > string(names[i]) {
			t.Errorf("List not sorted: %q > %q", names[i-1], names[i])
		}
	}
}

func TestLoader_ResolvePath(t *testing.T) {
	l := NewLoader("/tmp/prompts")
	path, err := l.ResolvePath(PromptMatchAnalysis)
	if err != nil {
		t.Fatalf("ResolvePath error: %v", err)
	}
	expected := "/tmp/prompts/match_analysis_v1.md"
	if path != expected {
		t.Errorf("ResolvePath = %q, want %q", path, expected)
	}
}

func TestLoader_ResolvePath_UnknownName(t *testing.T) {
	l := NewLoader("/tmp/prompts")
	_, err := l.ResolvePath(PromptName("nonexistent"))
	if err == nil {
		t.Fatal("expected error for unknown prompt name, got nil")
	}
}

func TestResolveDir_ExistingPath(t *testing.T) {
	// testBaseDir is a known existing directory.
	result := resolveDir(testBaseDir)
	if result != testBaseDir {
		t.Errorf("resolveDir(existing) = %q, want %q", result, testBaseDir)
	}
}

func TestResolveDir_ServerPrefixStrip(t *testing.T) {
	// When cwd is the server/ directory, "server/prompts" should resolve to "prompts".
	// We simulate this by creating a temp dir structure.
	tmpDir := t.TempDir()
	serverDir := filepath.Join(tmpDir, "server", "prompts")
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Save and restore cwd.
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Change to the "server" subdirectory.
	if err := os.Chdir(filepath.Join(tmpDir, "server")); err != nil {
		t.Fatal(err)
	}

	// "server/prompts" doesn't exist from server/ dir, but "prompts" does.
	result := resolveDir("server/prompts")
	expected := "prompts"
	if result != expected {
		t.Errorf("resolveDir(\"server/prompts\") from server/ = %q, want %q", result, expected)
	}
}

func TestResolveDir_NonExistent(t *testing.T) {
	// Non-existent paths that don't match "server/" prefix should return as-is.
	result := resolveDir("/nonexistent/prompts")
	if result != "/nonexistent/prompts" {
		t.Errorf("resolveDir(nonexistent) = %q, want %q", result, "/nonexistent/prompts")
	}
}

func TestNewLoader_ResolvesServerPrefix(t *testing.T) {
	// When cwd is server/ and baseDir is "server/prompts", NewLoader should
	// resolve to "prompts" automatically.
	tmpDir := t.TempDir()
	serverDir := filepath.Join(tmpDir, "server", "prompts")
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(filepath.Join(tmpDir, "server")); err != nil {
		t.Fatal(err)
	}

	l := NewLoader("server/prompts")
	// The loader should have resolved to "prompts" and be able to load.
	// (It won't find actual prompt files in the temp dir, but baseDir should be correct.)
	if l.baseDir != "prompts" {
		t.Errorf("NewLoader(\"server/prompts\").baseDir = %q, want %q", l.baseDir, "prompts")
	}
}
