package main

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/buildkite/terminal-to-html/v3"
	"io"
	"maze.io/x/ttyrec"
	"os"
	"text/template"
	"time"

	"github.com/alecthomas/kong"
)

type Config struct {
	Date       bool   `help:"Show date" short:"d" default:"false"`
	Html       bool   `help:"Display result in HTML" short:"H" default:"false"`
	RecordFile string `help:"Dashboard page" default:"all" arg:""`
}

//go:embed record.html.tmpl
var html_template embed.FS

type TemplateData struct {
	Config Config
	Frames []Frame
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

	records := m.ReadAll(config)
	if config.Html {
		tmpl, err := html_template.ReadFile("record.html.tmpl")
		if err != nil {
			fatalf("error loading HTML template: %v\n", err)
		}
		t := template.Must(template.New("HTML Record").Funcs(template.FuncMap{"FormatDate": DateFormater(config), "Ansi2Html": terminal.Render}).Parse(string(tmpl)))
		if err != nil {
			fatalf("error loading HTML template: %v\n", err)
		}

		templateData := TemplateData{
			Config: config,
			Frames: records,
		}
		err = t.Execute(os.Stdout, templateData)
		if err != nil {
			fatalf("error rendering HTML template: %v\n", err)
		}
	} else {
		var result [][]byte
		for _, frame := range records {
			if config.Date {
				content := append([]byte(FormatDate(frame.Date)+": "), frame.Data...)
				result = append(result, content)
			} else {
				result = append(result, frame.Data)
			}
		}
		_, err := io.WriteString(os.Stdout, string(bytes.Join(result, []byte("\n"))))
		if err != nil {
			fatalf("writing to stdout: %v\n", err)
		}
	}
}

func DateFormater(config Config) func(time.Time) string {
	if !config.Date {
		return func(time.Time) string {
			return ""
		}
	}
	return func(date time.Time) string {
		return FormatDate(date)
	}
}

func FormatDate(date time.Time) string {
	return date.Format("2006-01-02 15:04:05")
}

func fatalf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
