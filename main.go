package main

import (
	_ "github.com/joy999/llama-proxy/internal/packed"

	_ "github.com/joy999/llama-proxy/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"github.com/joy999/llama-proxy/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
