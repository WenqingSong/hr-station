package main

import (
	"fmt"
	"os"

	// Register the MySQL driver for the GoFrame ORM.
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/joho/godotenv"

	"offer-catcher/agents/internal/cmd"
)

func main() {
	// Load .env for local development (agents/ first, then parent project dirs).
	// godotenv.Load does not override already-set OS environment variables.
	_ = godotenv.Load(".env", "../.env", "../../.env")

	setDefaultEnv()

	// 校验必填数据库环境变量并初始化 gdb 配置；失败则输出明确错误并终止。
	if err := setupDatabase(); err != nil {
		fmt.Fprintf(os.Stderr, "[agents] fatal: %v\n", err)
		os.Exit(1)
	}

	cmd.Main.Run(gctx.GetInitCtx())
}

// setDefaultEnv fills only non-sensitive environment variables with safe
// defaults. Database credentials are intentionally NOT defaulted here.
func setDefaultEnv() {
	if os.Getenv("AGENTS_PORT") == "" {
		_ = os.Setenv("AGENTS_PORT", "18080")
	}
	if os.Getenv("AGENTS_DB_NAME") == "" {
		_ = os.Setenv("AGENTS_DB_NAME", "offer_catcher_re")
	}
}

// setupDatabase validates the required database environment variables and
// initializes the global gdb configuration. It returns an error (rather than
// panicking or continuing) when any required variable is missing.
func setupDatabase() error {
	required := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD"}
	var missing []string
	for _, k := range required {
		if os.Getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required database environment variables: %v (see .env.example)", missing)
	}

	link := fmt.Sprintf(
		"mysql:%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("AGENTS_DB_NAME"),
	)

	return gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{
			gdb.ConfigNode{
				Link:  link,
				Debug: true,
			},
		},
	})
}
