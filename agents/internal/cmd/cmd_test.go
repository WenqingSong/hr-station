package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"offer-catcher/agents/internal/controller/health"
	"offer-catcher/agents/internal/middleware"
)

const testAddress = "127.0.0.1:18123"

// TestMain sets up a database configuration for the in-process server tests so
// that g.DB() does not fail with a missing-configuration error. The link points
// to an unreachable address, so the health check deterministically reports
// database status "unavailable" without depending on an external MySQL.
func TestMain(m *testing.M) {
	_ = gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{
			gdb.ConfigNode{
				Link:  "mysql:root:unused@tcp(127.0.0.1:1)/test?charset=utf8mb4&parseTime=True&loc=Local&timeout=2s",
				Debug: false,
			},
		},
	})
	os.Exit(m.Run())
}

// TestHealthzEndpoint is an automated interface test that starts the GoFrame
// server in-process and performs a real HTTP GET against /healthz.
func TestHealthzEndpoint(t *testing.T) {
	s := g.Server("agents-health-test")
	s.SetAddr(testAddress)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Middleware(middleware.DBHealth)
		group.Bind(
			health.NewV1(),
		)
	})

	if err := s.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer s.Shutdown()

	url := "http://" + testAddress + "/healthz"
	client := &http.Client{Timeout: 5 * time.Second}

	// Wait for the server to become ready (it starts asynchronously).
	var resp *http.Response
	var err error
	for i := 0; i < 30; i++ {
		resp, err = client.Get(url)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Service  string `json:"service"`
			Env      string `json:"env"`
			App      string `json:"app"`
			Database string `json:"database"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Code != 0 {
		t.Errorf("code = %d, want 0", body.Code)
	}
	if body.Data.Service == "" {
		t.Errorf("data.service is empty")
	}
	if body.Data.App != "ok" {
		t.Errorf("data.app = %q, want ok", body.Data.App)
	}
	if body.Data.Database != "ok" && body.Data.Database != "unavailable" {
		t.Errorf("data.database = %q, want ok or unavailable", body.Data.Database)
	}
}
