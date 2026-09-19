package v1

import "github.com/gogf/gf/v2/frame/g"

// HealthReq is the request for the health check endpoint.
type HealthReq struct {
	g.Meta `path:"/healthz" method:"get" tags:"Health" summary:"Health check"`
}

// HealthRes is the response for the health check endpoint.
type HealthRes struct {
	Service  string `json:"service"  dc:"服务名称"`
	Env      string `json:"env"      dc:"运行环境"`
	App      string `json:"app"      dc:"应用状态"`
	Database string `json:"database" dc:"数据库状态"`
}
