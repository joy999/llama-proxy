package v1

import (
	"llama-proxy/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// Model 模型信息结构体
// dc:"AI模型基本信息，符合OpenAI API规范"
type Model = model.Model

// ListModelsReq 获取模型列表请求
// dc:"获取所有可用模型的列表，符合OpenAI API规范"
type ListModelsReq struct {
	g.Meta `path:"/models" method:"get" tags:"OpenAI Models" summary:"获取模型列表"`
}

// ListModelsRes 获取模型列表响应
// dc:"模型列表响应结果，符合OpenAI API规范"
type ListModelsRes struct {
	g.Meta `mime:"application/json"`
	Object string   `json:"object" dc:"对象类型，固定为list"`
	Data   []*Model `json:"data" dc:"模型列表数据"`
}

// GetModelReq 获取单个模型详情请求
// dc:"根据模型ID获取模型的详细信息，符合OpenAI API规范"
type GetModelReq struct {
	g.Meta  `path:"/models/{model}" method:"get" tags:"OpenAI Models" summary:"获取模型详情"`
	ModelId string `json:"model" v:"required" dc:"模型ID"`
}

// GetModelRes 获取单个模型详情响应
// dc:"模型详情响应结果，符合OpenAI API规范"
type GetModelRes struct {
	g.Meta `mime:"application/json"`
	*Model
}
