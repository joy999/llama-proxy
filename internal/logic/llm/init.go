package llm

import "llama-proxy/internal/service"

func init() {
	service.RegisterLLM(newLLM())
}
