package cmd

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gogf/gf/v2/os/gcfg"
)

// TestConfigLoads verifies that manifest/config/config.yaml is parsed and the
// key non-sensitive configuration paths (app / server) are resolvable.
func TestConfigLoads(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	configDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "manifest", "config")

	adapter, err := gcfg.NewAdapterFile()
	if err != nil {
		t.Fatalf("create config adapter: %v", err)
	}
	if err := adapter.SetPath(configDir); err != nil {
		t.Fatalf("set config path: %v", err)
	}

	cfg, err := gcfg.New()
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	cfg.SetAdapter(adapter)
	ctx := context.Background()

	if got := cfg.MustGet(ctx, "app.name").String(); got != "offer-catcher-agents" {
		t.Errorf("app.name = %q, want offer-catcher-agents", got)
	}
	if got := cfg.MustGet(ctx, "server.address").String(); got != ":18080" {
		t.Errorf("server.address = %q, want :18080", got)
	}
}
