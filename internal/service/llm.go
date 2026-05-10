// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"net/http"

	"github.com/joy999/llama-proxy/internal/model"
)

type (
	ILLM interface {
		ModelList(ctx context.Context) []*model.Model
		ModelDetail(ctx context.Context, modelId string) (*model.Model, error)
		LoadModel(ctx context.Context, modelId string) (addr string, err error)
		UnloadModel(ctx context.Context, modelId string) (err error)
		IsModelLoaded(ctx context.Context, modelId string) (bool, error)
		ServeHTTP(ctx context.Context, r *http.Request, w http.ResponseWriter)
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
