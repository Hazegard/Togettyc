package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"maze.io/x/ttyrec"
	"os"
	"os/exec"
	"strings"
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
	TmuxMode   bool      `help:"Tmux" short:"T" default:"false" hidden:"true"`
}

func (c *Config) WriteCli() string {
	args := []string{}
	if c.Date {
		args = append(args, "--date")
	}
	if c.NoColor {
		args = append(args, "--no-color")
	}
	if c.Html {
		args = append(args, "--html")
	}
	if c.RecordFile != "" {
		args = append(args, c.RecordFile)
	}
	return strings.Join(args, " ")
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

	records := []Frame{}
	if config.TmuxMode {
		records = m.ReadAll()
	} else {
		records, err = tmux(config)
	}
	if err != nil {
		fatalf("error reading records: %v\n", err)
	}
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

func tmux(config Config) ([]Frame, error) {
	fmt.Println("tmux")
	fmt.Println("tmux")
	fmt.Println("tmux")
	fmt.Println("tmux")
	fmt.Println("tmux")
	fmt.Println("tmux")
	fmt.Println("tmux")
	fmt.Println("tmux")
	fmt.Println("tmux")
	fmt.Println("tmux")
	exePath, err := os.Executable()
	if err != nil {
		return []Frame{}, err
	}
	id := randString(16)
	args := []string{
		"new-session",
		"-ds",
		id,
		"\\;",
		"send-keys",
		"-t",
		fmt.Sprintf("%s:1.0", id),
		fmt.Sprintf("%s -T %s; tmux wait -S %s", exePath, config.WriteCli(), id),
		"C-m",
		"\\;",
		"wait",
		id,
		"\\;",
		"capture-pane",
		"-t",
		fmt.Sprintf("%s:1.0", id),
		"-eCpNJ",
		"-S",
		"-",
		"-E",
		"-",
		"\\;",
		"kill-session",
		"-t",
		id,
	}
	cmd := exec.Command("tmux", args...)
	fmt.Println(cmd.Args)

	var out bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Run the command.
	if err := cmd.Run(); err != nil {
		return []Frame{}, err
	}
	fmt.Println(len(out.Bytes()))
	fmt.Println(len(out.Bytes()))
	fmt.Println(len(out.Bytes()))
	fmt.Println(len(out.Bytes()))
	fmt.Println(len(out.Bytes()))
	fmt.Println(len(out.Bytes()))

	// res := out.String()
	newRecords := []Frame{}
	date := []byte{}
	for i, line := range bytes.Split(out.Bytes(), []byte("\n")) {
		if i%2 == 0 && config.Date {
			date = line
		} else {
			date, err := time.ParseInLocation(timeFormat, string(date), time.Local)
			if err != nil {
				fatalf("error parsing date: %v\n", err)
				continue
			}
			records := Frame{
				Data: line,
				Date: date,
			}
			newRecords = append(newRecords, records)
		}
	}
	fmt.Println(len(newRecords))
	return newRecords, nil
}

// tmux new-session -ds TEST \; send-keys -t TEST:1.0 "go run . -d --no-color  output.log.ttyrec; tmux wait -S ping" C-m \; wait ping \; capture-pane -t TEST:1.0 -eCpNJ -S - -E - \; kill-session -t TEST
func IsInPath(cmd string) bool {
	res, err := exec.LookPath(cmd)
	if err != nil {
		return false
	}
	return res != ""
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	// Seed the random number generator.
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
