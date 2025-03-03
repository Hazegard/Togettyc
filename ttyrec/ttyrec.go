package ttyrec

import (
	"fmt"
	"github.com/klauspost/compress/zstd"
	"io"
	"os"
	"strings"
	"time"
)

type Config struct {
	Append   bool     `help:"append to the existing file" short:"a"`
	Compress bool     `help:"compress to the output file" short:"Z"`
	Output   string   `help:"output file name" short:"f"`
	Shell    string   `help:"shell to use, using current one by default" short:"S"`
	Exe      string   `help:"Command to execute" arg:"" optional""`
	Args     []string `arg:"" help:"arguments" optional""`
}

func (cfg *Config) Run() error {
	if cfg.Output == "" {
		cfg.Output = fmt.Sprintf("togettyc-%s.log", time.Now().Format("20060102150405"))
	}
	if cfg.Compress && !strings.HasSuffix(cfg.Output, ".zst") {
		cfg.Output += ".zst"
	}
	return run(*cfg)
}

func openEncoder(config Config) (io.WriteCloser, error) {
	var (
		f   *os.File
		err error
	)

	if config.Append {
		f, err = os.OpenFile(config.Output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	} else {
		f, err = os.Create(config.Output)
	}
	if err != nil {
		return nil, err
	}

	if config.Compress {
		zstdFile, err := zstd.NewWriter(f)
		if err != nil {
			return nil, err
		}
		return zstdFile, nil
	} else {
		return f, nil
	}
}

func fatalf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
