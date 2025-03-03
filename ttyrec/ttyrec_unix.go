//go:build linux || darwin

package ttyrec

import (
	"fmt"
	"github.com/creack/pty"
	"io"
	"log"
	"maze.io/x/ttyrec"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
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

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("error starting command in pty: %v", err)
	}

	defer func() { _ = ptmx.Close() }()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT)
		for range sigCh {
			// Send SIGINT to the entire process group of the child.
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT); err != nil {
				log.Printf("failed to forward SIGINT: %v", err)
			}
		}
	}()

	// Resize the pty to match the current terminal size.
	if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
		return fmt.Errorf("error resizing pty: %v", err)
	}

	// Set up a signal handler to watch for window size changes.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		for range sigCh {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %v", err)
			}
		}
	}()

	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	// cmd.Stdin = os.Stdin
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	return fmt.Errorf("error opening stdout pipe: %v", err)
	// }
	// defer stdout.Close()
	// stderr, err := cmd.StderrPipe()
	// if err != nil {
	// 	return fmt.Errorf("error opening stderr pipe: %v", err)
	// }
	// defer stderr.Close()

	go func(r io.Reader) {
		f, err := openEncoder(config)
		if err != nil {
			fatalf("error opening encoder: %v", err)
		}
		defer f.Close()
		e := ttyrec.NewEncoder(f)

		//if _, err := io.Copy(e, r); err != nil {
		//	fatalf("error writing to encoder: %v", err)
		//}
		_, err = io.Copy(io.MultiWriter(e, os.Stdout), r)
		if err != nil {
			fatalf("error writing to encoder: %v", err)
		}
	}( /*io.MultiReader(
			io.TeeReader(stdout, os.Stdout),
			io.TeeReader(stderr, os.Stderr),
		)) */ptmx)

	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("error running command: %v", err)
	}
	return nil
}
