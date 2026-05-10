package openai

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	v1 "llama-proxy/api/openai/v1"
	"llama-proxy/internal/service"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

func (c *Controller) ListModels(ctx context.Context, req *v1.ListModelsReq) (res *v1.ListModelsRes, err error) {
	res = &v1.ListModelsRes{
		Object: "list",
		Data:   service.LLM().ModelList(ctx),
	}
	return
}

func (c *Controller) GetModel(ctx context.Context, req *v1.GetModelReq) (res *v1.GetModelRes, err error) {
	model, err := service.LLM().ModelDetail(ctx, req.ModelId)
	if err != nil {
		return nil, err
	}
	res = &v1.GetModelRes{
		Model: model,
	}
	return
}

func ProxyHandler(r *ghttp.Request) {
	service.LLM().ServeHTTP(r.Context(), r)
}
