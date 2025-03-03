package ttyprint

import (
	"embed"
	"github.com/buildkite/terminal-to-html/v3"
	"os"
	"text/template"
	"time"
)

//go:embed record.html.tmpl
var html_template embed.FS

type TemplateData struct {
	Config Config
	Frames []Frame
}

func GenerateHtml(config Config, records []Frame) {
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
