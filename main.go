package main

import (
	"github.com/alecthomas/kong"
	"log"
	"os"
	"path/filepath"
	"togettyc/ttyplay"
	"togettyc/ttyprint"
	"togettyc/ttyrec"
	"togettyc/ttytime"
)

type Config struct {
	Play  ttyplay.Config  `cmd:"" help:"Play the record"`
	Print ttyprint.Config `cmd:"" help:"Render the record"`
	Rec   ttyrec.Config   `cmd:"" help:"Record" default:"withargs"`
	Time  ttytime.Config  `cmd:"" help:"Time to record"`
}

func (c *Config) Run() error {
	return nil
}

func main() {

	name := filepath.Base(os.Args[0])
	var ctx *kong.Context
	kongOptions := []kong.Option{
		kong.Name(name),
		kong.Description("Cross-platform reimplementation of ttyrec"),
		kong.UsageOnError(),
	}
	switch name {
	case "ttyplay":
		ctx = kong.Parse(&ttyplay.Config{}, kongOptions...)
	case "ttyrec":
		ctx = kong.Parse(&ttyrec.Config{}, kongOptions...)
	case "ttyprint":
		ctx = kong.Parse(&ttyprint.Config{}, kongOptions...)
	case "ttytime":
		ctx = kong.Parse(&ttytime.Config{}, kongOptions...)
	default:
		ctx = kong.Parse(&Config{}, kongOptions...)
	}

	err := ctx.Run()
	if err != nil {
		log.Fatal(err)
	}
}
