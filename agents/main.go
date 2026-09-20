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
	// 只加载 agents/.env（单文件）。.env.example 已自带可用配置（含本地密码），
	// 适配 CNB 容器每天空白重启场景。
	_ = godotenv.Load(".env")

	setupDatabase()

	cmd.Main.Run(gctx.GetInitCtx())
}

// setupDatabase 注入数据库连接配置（DSN 由 .env 提供，不做缺失校验）。
func setupDatabase() {
	link := fmt.Sprintf(
		"mysql:%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("AGENTS_DB_NAME"),
	)

	_ = gdb.SetConfig(gdb.Config{
		gdb.DefaultGroupName: gdb.ConfigGroup{
			gdb.ConfigNode{
				Link:  link,
				Debug: true,
			},
		},
	})
}
