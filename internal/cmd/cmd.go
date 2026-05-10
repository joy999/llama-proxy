package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"github.com/joy999/llama-proxy/internal/controller/openai"
	"github.com/joy999/llama-proxy/internal/middleware"
)

var (
	Main = gcmd.Command{
		Name:  "LLamaProxy",
		Usage: "LLamaProxy",
		Brief: "start LLamaProxy server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/v1", func(group *ghttp.RouterGroup) {
				group.Group("/", func(standardGroup *ghttp.RouterGroup) {
					standardGroup.Middleware(ghttp.MiddlewareHandlerResponse, middleware.ResponseHandler)
					standardGroup.Bind(openai.New())
				})

				group.ALL("*", openai.ProxyHandler)
			})
			s.Run()
			return nil
		},
	}
)
