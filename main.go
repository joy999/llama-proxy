package main

import (
	_ "llama-proxy/internal/packed"

	_ "llama-proxy/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"llama-proxy/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
