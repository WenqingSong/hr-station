package schema

import (
	"os"
	"path/filepath"
	"testing"
)

// testBaseDir points to the real schema directory for integration tests.
var testBaseDir = func() string {
	d, _ := filepath.Abs("../../prompts/schema")
	return d
}()

func TestLoader_Load_MatchAnalysis(t *testing.T) {
	l := NewLoader(testBaseDir)
	doc, err := l.Load(SchemaMatchAnalysis)
	if err != nil {
		t.Fatalf("Load(match_analysis) error: %v", err)
	}
	if doc.Name != SchemaMatchAnalysis {
		t.Errorf("Name = %q, want %q", doc.Name, SchemaMatchAnalysis)
	}
	if doc.Content == "" {
		t.Error("Content is empty")
	}
	if doc.Raw == nil {
		t.Error("Raw is nil")
	}
}

func TestLoader_Load_Recommendation(t *testing.T) {
	l := NewLoader(testBaseDir)
	doc, err := l.Load(SchemaRecommendation)
	if err != nil {
		t.Fatalf("Load(recommendation) error: %v", err)
	}
	if doc.Content == "" {
		t.Error("Content is empty")
	}
}

func TestLoader_Load_ResumeOptimization(t *testing.T) {
	l := NewLoader(testBaseDir)
	doc, err := l.Load(SchemaResumeOptimization)
	if err != nil {
		t.Fatalf("Load(resume_optimization) error: %v", err)
	}
	if doc.Content == "" {
		t.Error("Content is empty")
	}
}

func TestLoader_Load_AllSchemasAreValidJSON(t *testing.T) {
	l := NewLoader(testBaseDir)
	for _, name := range l.List() {
		doc, err := l.Load(name)
		if err != nil {
			t.Errorf("Load(%s) error: %v", name, err)
			continue
		}
		if doc.Raw == nil {
			t.Errorf("Load(%s): Raw is nil, expected parsed JSON object", name)
		}
	}
}

func TestLoader_Load_Title(t *testing.T) {
	l := NewLoader(testBaseDir)
	for _, name := range l.List() {
		doc, err := l.Load(name)
		if err != nil {
			t.Fatalf("Load(%s) error: %v", name, err)
		}
		if doc.Title() == "" {
			t.Errorf("Schema %s: Title() is empty, expected non-empty title", name)
		}
	}
}

func TestLoader_Load_RequiredCount(t *testing.T) {
	l := NewLoader(testBaseDir)

	// match_analysis_schema.json has 17 required fields.
	doc, err := l.Load(SchemaMatchAnalysis)
	if err != nil {
		t.Fatalf("Load(match_analysis) error: %v", err)
	}
	if doc.RequiredCount() == 0 {
		t.Error("RequiredCount() = 0, expected > 0")
	}
}

func TestLoader_Load_UnknownName(t *testing.T) {
	l := NewLoader(testBaseDir)
	_, err := l.Load(SchemaName("nonexistent"))
	if err == nil {
		t.Fatal("expected error for unknown schema name, got nil")
	}
}

func TestLoader_List(t *testing.T) {
	l := NewLoader(testBaseDir)
	names := l.List()

	if len(names) != 3 {
		t.Fatalf("List returned %d names, want 3", len(names))
	}

	expected := map[SchemaName]bool{
		SchemaMatchAnalysis:      true,
		SchemaRecommendation:     true,
		SchemaResumeOptimization: true,
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

func TestLoader_Load_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "bad_schema.json"), []byte("not json"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	testName := SchemaName("bad_schema")
	mapping[testName] = "bad_schema.json"
	defer delete(mapping, testName)

	l := NewLoader(dir)
	_, err = l.Load(testName)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoader_ResolvePath_UnknownName(t *testing.T) {
	l := NewLoader("/tmp/schema")
	_, err := l.ResolvePath(SchemaName("nonexistent"))
	if err == nil {
		t.Fatal("expected error for unknown schema name, got nil")
	}
}

func TestDocument_Draft(t *testing.T) {
	l := NewLoader(testBaseDir)
	doc, err := l.Load(SchemaMatchAnalysis)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	draft := doc.Draft()
	if draft == "" {
		t.Error("Draft() is empty, expected $schema URI")
	}
}

func TestResolveDir_ExistingPath(t *testing.T) {
	result := resolveDir(testBaseDir)
	if result != testBaseDir {
		t.Errorf("resolveDir(existing) = %q, want %q", result, testBaseDir)
	}
}

func TestResolveDir_ServerPrefixStrip(t *testing.T) {
	tmpDir := t.TempDir()
	schemaDir := filepath.Join(tmpDir, "server", "prompts", "schema")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(filepath.Join(tmpDir, "server")); err != nil {
		t.Fatal(err)
	}

	result := resolveDir("server/prompts/schema")
	expected := "prompts/schema"
	if result != expected {
		t.Errorf("resolveDir(\"server/prompts/schema\") from server/ = %q, want %q", result, expected)
	}
}

func TestResolveDir_NonExistent(t *testing.T) {
	result := resolveDir("/nonexistent/schema")
	if result != "/nonexistent/schema" {
		t.Errorf("resolveDir(nonexistent) = %q, want %q", result, "/nonexistent/schema")
	}
}

func TestNewLoader_ResolvesServerPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	schemaDir := filepath.Join(tmpDir, "server", "prompts", "schema")
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(filepath.Join(tmpDir, "server")); err != nil {
		t.Fatal(err)
	}

	l := NewLoader("server/prompts/schema")
	if l.baseDir != "prompts/schema" {
		t.Errorf("NewLoader(\"server/prompts/schema\").baseDir = %q, want %q", l.baseDir, "prompts/schema")
	}
}
