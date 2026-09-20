package health

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "offer-catcher/agents/api/health/v1"
)

// ControllerV1 implements the health check controller.
type ControllerV1 struct{}

// NewV1 returns a new health check controller instance.
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// Health handles GET /healthz and reports application + database status.
func (c *ControllerV1) Health(ctx context.Context, req *v1.HealthReq) (res *v1.HealthRes, err error) {
	database := "ok"
	if err := g.DB().PingMaster(); err != nil {
		database = "unavailable"
	}

	res = &v1.HealthRes{
		Service:  g.Cfg().MustGet(ctx, "app.name", "offer-catcher-agents").String(),
		Env:      g.Cfg().MustGet(ctx, "app.env", "local").String(),
		App:      "ok",
		Database: database,
	}

	return res, nil
}
