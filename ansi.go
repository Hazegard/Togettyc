package main

import (
	"bytes"
	"io"
	"os"
	"regexp"
)

const ansi2 = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
const ansi = `\x1B\[[0-9;]*[a-zA-Z]`

var re = regexp.MustCompile(ansi)

func StripAll(records []Frame) []Frame {
	var strippedFrames []Frame
	for _, record := range records {
		strippedData := []byte(Strip(string(record.Data)))

		strippedFrame := Frame{
			Date: record.Date,
			Data: strippedData,
		}
		strippedFrames = append(strippedFrames, strippedFrame)
	}
	return strippedFrames
}

func Strip(str string) string {
	return re.ReplaceAllString(str, "")
}

func GenerateConsoleOutput(config Config, records []Frame) {
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
