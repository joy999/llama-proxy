// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"llama-proxy/internal/model"

	"github.com/gogf/gf/v2/net/ghttp"
)

type (
	ILLM interface {
		// ModelList 获取模型列表
		ModelList(ctx context.Context) []*model.Model
		// ModelDetail 获取模型详情
		ModelDetail(ctx context.Context, modelId string) (*model.Model, error)
		// LoadModel 加载模型
		LoadModel(ctx context.Context, modelId string) (addr string, err error)
		// UnloadModel 卸载模型
		UnloadModel(ctx context.Context, modelId string) (err error)
		// IsModelLoaded 判定模型是否已经加载
		IsModelLoaded(ctx context.Context, modelId string) (bool, error)
		// ServeHTTP 接口反代
		ServeHTTP(ctx context.Context, r *ghttp.Request)
	}
)

var (
	localLLM ILLM
)

func LLM() ILLM {
	if localLLM == nil {
		panic("implement not found for interface ILLM, forgot register?")
	}
	return localLLM
}

func RegisterLLM(i ILLM) {
	localLLM = i
}
