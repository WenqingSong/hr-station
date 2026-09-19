package cmd

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"

	"offer-catcher/agents/internal/controller/health"
	"offer-catcher/agents/internal/middleware"
)

// Main is the entry command for the agents HTTP server.
var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "offer-catcher agents HTTP server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 1) 配置由 GoFrame 在启动时自动加载（manifest/config/config.yaml）。
			appName := g.Cfg().MustGet(ctx, "app.name", "offer-catcher-agents").String()
			appEnv := g.Cfg().MustGet(ctx, "app.env", "local").String()
			address := g.Cfg().MustGet(ctx, "server.address", ":18080").String()
			glog.Infof(ctx, "[agents] config loaded: app=%s env=%s address=%s", appName, appEnv, address)

			// 2) 启动前检查数据库连接（限时 Ping，失败即退出，不监听端口）。
			if err := checkDatabase(ctx); err != nil {
				glog.Fatalf(ctx, "[agents] database connectivity check failed: %v", err)
				return err
			}

			// 3) 注册路由（数据库健康中间件仅用于运行期间报告状态）。
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Middleware(middleware.DBHealth)
				group.Bind(
					health.NewV1(),
				)
			})

			// 4) 启动 HTTP 服务（阻塞；GoFrame 内置优雅退出）。
			s.Run()
			return nil
		},
	}
)

// dbPingTimeout bounds the startup database connectivity check.
const dbPingTimeout = 5 * time.Second

// checkDatabase verifies the default database connection is reachable with a
// bounded timeout. It runs on every startup and fails fast when unavailable,
// so the HTTP server never starts listening when the database is down.
func checkDatabase(ctx context.Context) error {
	if err := pingDatabase(dbPingTimeout); err != nil {
		return gerror.Wrapf(err, "ping mysql failed (check DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/AGENTS_DB_NAME)")
	}

	glog.Infof(ctx, "[agents] database connected")
	return nil
}

// pingDatabase pings the default database with a hard timeout so a slow or
// unreachable host cannot block startup indefinitely.
func pingDatabase(timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- g.DB().PingMaster()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return gerror.New("ping mysql timeout")
	}
}
