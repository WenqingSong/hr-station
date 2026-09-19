// Package middleware 提供 Agents 服务的 HTTP 中间件。
package middleware

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ctxKeyDBStatus 是请求上下文中存储数据库健康状态的键。
// 使用独立类型避免与其他包的 context key 冲突。
type ctxKey string

const ctxKeyDBStatus ctxKey = "agents.db.status"

// DBHealth 数据库健康检查中间件。
//
// 对默认数据库执行一次 PingMaster 检查，并把结果（"ok" / "unavailable"）
// 写入请求上下文，供后续控制器（如 /healthz）读取。
// 该中间件不中断请求，仅负责探测数据库连通性。
func DBHealth(r *ghttp.Request) {
	status := "ok"
	if err := g.DB().PingMaster(); err != nil {
		status = "unavailable"
	}

	r.SetCtxVar(ctxKeyDBStatus, status)
	r.Middleware.Next()
}

// GetDBStatus 从请求上下文中读取 DBHealth 中间件写入的数据库健康状态。
// 若上下文中不存在该值（例如中间件未挂载），默认返回 "ok"。
func GetDBStatus(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyDBStatus).(string); ok {
		return v
	}
	return "ok"
}
