package main

import (
	"github.com/alecthomas/kong"
	"log"
	"togettyc/ttyplay"
	"togettyc/ttyprint"
	"togettyc/ttyrec"
)

type Config struct {
	Print ttyprint.Config `cmd:"" help:"Render the record"`
	Rec   ttyrec.Config   `cmd:"" help:"Record" default:"withargs"`
	Play  ttyplay.Config  `cmd:"" help:"Play the record"`
}

func main() {
	cfg := Config{}
	kongOptions := []kong.Option{
		kong.Name("Togettyc"),
		kong.Description("Cross-platform reimplementation og ttyrec"),
		kong.UsageOnError(),
	}
	ctx := kong.Parse(&cfg, kongOptions...)
	err := ctx.Run()
	if err != nil {
		log.Fatal(err)
	}
}
