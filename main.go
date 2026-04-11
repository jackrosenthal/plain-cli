package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/jackrosenthal/plain-cli/cmd"
)

func main() {
	cli := cmd.CLI{}
	ctx := kong.Parse(&cli, kong.Name("plain"))
	if err := ctx.Run(); err != nil {
		ctx.FatalIfErrorf(err)
		os.Exit(1)
	}
}
