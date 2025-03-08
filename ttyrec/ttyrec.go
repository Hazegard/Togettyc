package ttyrec

import (
	"fmt"
	"github.com/klauspost/compress/zstd"
	"github.com/runletapp/go-console"
	"golang.org/x/term"
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

func run(config Config) error {

	if config.Shell == "" {
		config.Shell = getShellCmd(append([]string{detectShell()}, shells...))
	}

	config.Args = formatArgs(config.Args)
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	proc, err := console.New(w, h)
	if err != nil {
		return err
	}
	err = proc.SetENV(os.Environ())
	if err != nil {
		return err
	}
	err = proc.Start(append([]string{config.Shell}, config.Args...))
	if err != nil {
		return fmt.Errorf("error starting command in pty: %v", err)
	}

	go handleResize(proc)

	defer func() { _ = proc.Close() }()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("error creating terminal state: %v", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() {
		_, _ = io.Copy(proc, os.Stdin)
	}()

	f, err := openEncoder(config)
	if err != nil {
		fatalf("error opening encoder: %v", err)
	}
	defer f.Close()
	e := NewEncoder(f)
	go func() {
		_, err = io.Copy(io.MultiWriter(e, os.Stdout), proc)
		if err != nil {
			fatalf("error writing to encoder: %v", err)
		}
	}()

	if _, err := proc.Wait(); err != nil {
		return fmt.Errorf("error running command: %v", err)
	}

	return nil
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
