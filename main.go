package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"maze.io/x/ttyrec"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/klauspost/compress/zstd"
)

type Config struct {
	Date       bool   `help:"Show date" short:"d" default:"false"`
	NoColor    bool   `help:"Disable colors" default:"false"`
	Html       bool   `help:"Display result in HTML" short:"H" default:"false"`
	RecordFile string `help:"Dashboard page" default:"all" arg:""`
}

// Zstandard magic bytes: 0x28, 0xB5, 0x2F, 0xFD.
var zstdMagic = []byte{0x28, 0xB5, 0x2F, 0xFD}

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
	// Wrap the original reader in a bufio.Reader to peek at magic bytes.
	br := bufio.NewReader(r)
	peek, err := br.Peek(4)
	if err != nil {
		fatalf("error reading magic bytes: %v\n", err)
	}

	if bytes.Equal(peek, zstdMagic) {
		// Create a zstd decoder using the buffered reader.
		decoder, err := zstd.NewReader(br)
		if err != nil {
			fatalf("error creating zstd reader: %v\n", err)
		}
		defer decoder.Close()
		r = decoder.IOReadCloser()
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
