package model

type Model struct {
	Id      string `json:"id" dc:"模型唯一标识符"`
	Object  string `json:"object" dc:"对象类型，固定为model"`
	Created int64  `json:"created,omitempty" dc:"模型创建时间戳"`
	OwnedBy string `json:"owned_by,omitempty" dc:"模型拥有者"`
}

type ModelConfig struct {
	Id              string  `json:"id"`
	Name            string  `json:"name"`
	OwnedBy         string  `json:"owned_by"`
	ModelPath       string  `json:"model_path"`
	VisionModelPath string  `json:"vision_model_path"`
	CtxSize         int     `json:"ctx_size"`
	Threads         int     `json:"threads"`
	Parallel        int     `json:"parallel"`
	GpuLayers       *int    `json:"gpu_layers"`
	CacheTypeK      string  `json:"cache_type_k"`
	CacheTypeV      string  `json:"cache_type_v"`
	FlashAttn       string  `json:"flash_attn"`
	NoMmap          bool    `json:"no_mmap"`
	Mlock           bool    `json:"mlock"`
	Temp            float64 `json:"temp"`
	TopP            float64 `json:"top_p"`
	TopK            int     `json:"top_k"`
	MinP            float64 `json:"min_p"`
	PresencePenalty float64 `json:"presence_penalty"`
	RepeatPenalty   float64 `json:"repeat_penalty"`
}

type DefaultParams struct {
	CtxSize         int     `json:"ctx_size"`
	Threads         int     `json:"threads"`
	Parallel        int     `json:"parallel"`
	GpuLayers       *int    `json:"gpu_layers"`
	CacheTypeK      string  `json:"cache_type_k"`
	CacheTypeV      string  `json:"cache_type_v"`
	FlashAttn       string  `json:"flash_attn"`
	VisionModelPath string  `json:"vision_model_path"`
	NoMmap          bool    `json:"no_mmap"`
	Mlock           bool    `json:"mlock"`
	Temp            float64 `json:"temp"`
	TopP            float64 `json:"top_p"`
	TopK            int     `json:"top_k"`
	MinP            float64 `json:"min_p"`
	PresencePenalty float64 `json:"presence_penalty"`
	RepeatPenalty   float64 `json:"repeat_penalty"`
}
