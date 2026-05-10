package llm

import "github.com/joy999/llama-proxy/internal/service"

func init() {
	service.RegisterLLM(newLLM())
}
