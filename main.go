package main

import (
	"fmt"
	"io"
	"maze.io/x/ttyrec"
	"os"
	"time"

	"github.com/alecthomas/kong"
)

type Config struct {
	Date       bool   `help:"Show date" short:"d" default:"false"`
	NoColor    bool   `help:"Disable colors" default:"false"`
	Html       bool   `help:"Display result in HTML" short:"H" default:"false"`
	RecordFile string `help:"Dashboard page" default:"all" arg:""`
}

func main() {
	config := Config{}
	_ = kong.Parse(&config)

	var r io.ReadCloser
	if config.RecordFile == "" {
		r = os.Stdin
	} else {
		var err error
		if r, err = os.Open(config.RecordFile); err != nil {
			fatalf("error opening %s: %v\n", config.RecordFile, err)
		}
	}
	defer r.Close()

	d := ttyrec.NewDecoder(r)

	m := Record{
		Decoder:      d,
		Frames:       []frame{},
		CurrentFrame: 0,
	}

	records := m.ReadAll()
	if config.NoColor {
		records = StripAll(records)
	}
	if config.Html {
		GenerateHtml(config, records)
	} else {
		GenerateConsoleOutput(config, records)
	}
}

func FormatDate(date time.Time) string {
	return date.Format("2006-01-02 15:04:05")
}

func fatalf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
