package llm

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	"llama-proxy/internal/service"
)

// LLM 是 llama-proxy 的导出层，通过 service 层访问内部功能
type LLM struct{}

// New 创建一个新的 LLM 实例
func New() *LLM {
	return &LLM{}
}

// GetModels 获取模型列表
func (l *LLM) GetModels(ctx context.Context) ([]map[string]interface{}, error) {
	models := service.LLM().ModelList(ctx)
	result := make([]map[string]interface{}, 0, len(models))
	for _, m := range models {
		result = append(result, map[string]interface{}{
			"id":     m.Id,
			"object": m.Object,
		})
	}
	return result, nil
}

// GetModel 获取单个模型
func (l *LLM) GetModel(ctx context.Context, modelId string) (map[string]interface{}, error) {
	model, err := service.LLM().ModelDetail(ctx, modelId)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, nil
	}
	return map[string]interface{}{
		"id":     model.Id,
		"object": model.Object,
	}, nil
}

// ServeHTTP 处理 HTTP 请求（GF 框架版本）
func (l *LLM) ServeHTTP(r *ghttp.Request) {
	l.ProxyHTTP(r.Context(), r.Request, r.Response.ResponseWriter)
}

// ProxyHTTP 处理 HTTP 请求（标准库版本）
func (l *LLM) ProxyHTTP(ctx context.Context, r *http.Request, w http.ResponseWriter) {
	service.LLM().ServeHTTP(ctx, r, w)
}

// LoadModel 加载模型
func (l *LLM) LoadModel(ctx context.Context, modelId string) (string, error) {
	return service.LLM().LoadModel(ctx, modelId)
}

// UnloadModel 卸载模型
func (l *LLM) UnloadModel(ctx context.Context, modelId string) error {
	return service.LLM().UnloadModel(ctx, modelId)
}

// IsModelLoaded 检查模型是否已加载
func (l *LLM) IsModelLoaded(ctx context.Context, modelId string) (bool, error) {
	return service.LLM().IsModelLoaded(ctx, modelId)
}
