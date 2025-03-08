//go:build linux || darwin

package ttyrec

import (
	"fmt"
	"github.com/creack/pty"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"strings"
)

func run(config Config) error {

	if config.Shell == "" {
		config.Shell = os.Getenv("SHELL")
		if config.Shell == "" {
			config.Shell = "/bin/sh"
		}
	}

	var cmd *exec.Cmd
	if len(config.Args) == 0 {
		cmd = exec.Command(config.Shell, "-i")
		// Set the process to run in its own process group.
		// cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	} else {
		cmd = exec.Command(config.Shell, "-ic", strings.Join(config.Args, " "))
		// Set the process to run in its own process group.
		// cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	cmd.Env = os.Environ()
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("error starting command in pty: %v", err)
	}

	defer func() { _ = ptmx.Close() }()

	// Resize the pty to match the current terminal size.
	if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
		return fmt.Errorf("error resizing pty: %v", err)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("error creating terminal state: %v", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	f, err := openEncoder(config)
	if err != nil {
		fatalf("error opening encoder: %v", err)
	}
	defer f.Close()
	e := NewEncoder(f)
	go func() {
		_, err = io.Copy(io.MultiWriter(e, os.Stdout), ptmx)
		if err != nil {
			fatalf("error writing to encoder: %v", err)
		}
	}()

	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("error running command: %v", err)
	}
	return nil
}
