package ttyprint

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
	"togettyc/ttyrec"
)

type Config struct {
	Date             bool      `help:"Show date" optional:"" short:"d" default:"false"`
	NoColor          bool      `help:"Disable colors" optional:"" default:"false"`
	Html             bool      `help:"Display result in HTML" optional:"" short:"H" default:"false"`
	RecordFile       string    `help:"Dashboard page" default:"all" arg:"" type:"existingfile"`
	StartDate        LocalTime `help:"Show results after the provided date (format:\"YYYY-MM-DD hh:mm:ss\")" optional:"" short:"S" `
	EndDate          LocalTime `help:"Show results before the provided date (format:\"YYYY-MM-DD hh:mm:ss\")" optional:"" short:"E" `
	Tmux             bool      `help:"Clean the output with tmux. It should reduce the noise provoked by garbage terminal manipulation" short:"T" default:"false"`
	InternalTmuxMode bool      `help:"Tmux" default:"false" hidden:"true"`
}

func (c *Config) WriteCli() string {
	args := []string{
		"--date",
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

func (c *Config) Run() error {

	var r io.ReadCloser
	if c.RecordFile == "" {
		r = os.Stdin
	} else {
		var err error
		if r, err = os.Open(c.RecordFile); err != nil {
			return fmt.Errorf("error opening %s: %v\n", c.RecordFile, err)
		}
	}
	// Wrap the original reader in a bufio.Reader to peek at magic bytes.
	br := bufio.NewReader(r)
	peek, err := br.Peek(4)
	if err != nil {
		return fmt.Errorf("error reading magic bytes: %v\n", err)
	}

	var decoder *ttyrec.Decoder
	if bytes.Equal(peek, zstdMagic) {
		// Create a zstd decoder using the buffered reader.
		zstdDecoder, err := zstd.NewReader(br)
		if err != nil {
			return fmt.Errorf("error creating zstd reader: %v\n", err)
		}
		originalFile := r
		defer originalFile.Close()
		defer zstdDecoder.Close()
		r = zstdDecoder.IOReadCloser()
		decoder = ttyrec.NewDecoder(r)
	} else {
		decoder = ttyrec.NewDecoder(br)
	}
	defer r.Close()

	m := Record{
		Decoder:      decoder,
		Frames:       []frame{},
		CurrentFrame: 0,
	}

	records := []Frame{}
	if c.InternalTmuxMode || !c.Tmux || !IsInPath("tmux") {
		records = m.ReadAll()
	} else {
		records, err = tmux(*c)
	}

	if err != nil {
		return fmt.Errorf("error reading records: %v\n", err)
	}
	if !c.EndDate.Get().IsZero() || !c.StartDate.Get().IsZero() {
		records = FilterRecordsDate(*c, records)
	}
	if c.NoColor {
		records = StripColor(records)
	}
	if c.Html {
		GenerateHtml(*c, records)
	} else {
		GenerateConsoleOutput(*c, records)
	}
	return nil
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
	exePath, err := os.Executable()
	if err != nil {
		return []Frame{}, err
	}
	id := randString(16)
	args := []string{
		"new-session",
		"-ds",
		id,
		";",
		"send-keys",
		"-t",
		fmt.Sprintf("%s:1.0", id),
		fmt.Sprintf(" %s print --internal-tmux-mode %s; tmux wait -S %s", exePath, config.WriteCli(), id),
		//fmt.Sprintf(" %s --internal-tmux-mode %s; tmux wait -S %s", exePath, config.WriteCli(), id),
		"C-m",
		";",
		"wait",
		id,
		";",
		"capture-pane",
		"-t",
		fmt.Sprintf("%s:1.0", id),
		"-eCpNJ",
		"-S",
		"-",
		"-E",
		"-",
		";",
		"kill-session",
		"-t",
		id,
	}
	cmd := exec.Command("tmux", args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	// Run the command.
	if err := cmd.Run(); err != nil {
		return []Frame{}, err
	}

	// res := out.String()
	newRecords := []Frame{}
	date := time.Time{}
	for _, line := range bytes.Split(out.Bytes(), []byte("\n")) {

		// We try to parse the current line as a date
		dd, err := time.ParseInLocation(timeFormat, string(StripColorBytes(bytes.ReplaceAll(line, []byte("\\033"), []byte("\033")))), time.Local)

		// IF succes, we hold it for the next iteration
		if err == nil {
			date = dd
			continue
		}

		// If the date is 0, skip
		if date.Format("2006") == "0001" {
			continue
		}
		// Remove ANSI escape sequences
		line = bytes.ReplaceAll(line, []byte("\\033"), []byte("\033"))
		line = bytes.ReplaceAll(line, []byte("\033[2J\033[3J\033[H"), []byte(""))
		line = bytes.ReplaceAll(line, []byte("\033[?1049h"), []byte(""))
		records := Frame{
			Data: line,
			Date: date,
		}
		newRecords = append(newRecords, records)

	}
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
