package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"offer-catcher/agents/internal/controller/health"
)

// Main is the entry command for the agents HTTP server.
var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "offer-catcher agents HTTP server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					health.NewV1(),
				)
			})

			s.Run()
			return nil
		},
	}
)
