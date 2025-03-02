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
	Date       bool      `help:"Show date" optional:"" short:"d" default:"false"`
	NoColor    bool      `help:"Disable colors" optional:"" default:"false"`
	Html       bool      `help:"Display result in HTML" optional:"" short:"H" default:"false"`
	RecordFile string    `help:"Dashboard page" default:"all" arg:"" type:"existingfile"`
	StartDate  LocalTime `help:"Show results after the provided date (format:\"YYYY-MM-DD hh:mm:ss\")" optional:"" short:"S" `
	EndDate    LocalTime `help:"Show results before the provided date (format:\"YYYY-MM-DD hh:mm:ss\")" optional:"" short:"E" `
}

type LocalTime time.Time

func (t *LocalTime) UnmarshalText(content []byte) error {
	tt, err := time.ParseInLocation(timeFormat, string(content), time.Local)
	if err != nil {
		return err
	}
	*t = LocalTime(tt)
	return nil
}

func (t *LocalTime) Get() time.Time {
	return time.Time(*t)
}

// Zstandard magic bytes: 0x28, 0xB5, 0x2F, 0xFD.
var (
	zstdMagic  = []byte{0x28, 0xB5, 0x2F, 0xFD}
	timeFormat = "2006-01-02 15:04:05"
)

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
		originalFile := r
		defer originalFile.Close()
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
	if !config.EndDate.Get().IsZero() || !config.StartDate.Get().IsZero() {
		records = FilterRecordsDate(config, records)
	}
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
	return date.Format(timeFormat)
}

func fatalf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}

func FilterRecordsDate(config Config, records []Frame) []Frame {
	newFrames := []Frame{}

	for _, frame := range records {
		if (config.EndDate.Get().IsZero() || frame.Date.Before(config.EndDate.Get())) && (config.StartDate.Get().IsZero() || frame.Date.After(config.StartDate.Get())) {
			newFrames = append(newFrames, frame)
		}
	}
	return newFrames
}
