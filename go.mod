module togettyc

go 1.24

require (
	github.com/alecthomas/kong v1.9.0
	github.com/buildkite/terminal-to-html/v3 v3.16.7
	github.com/klauspost/compress v1.18.0
	github.com/runletapp/go-console v0.0.0-20211204140000-27323a28410a
	golang.org/x/sys v0.31.0
	golang.org/x/term v0.30.0
	maze.io/x/ttyrec v1.0.0
)

require (
	github.com/creack/pty v1.1.24 // indirect
	github.com/iamacarpet/go-winpty v1.0.4 // indirect
)

replace (
	maze.io/x/ttyrec v1.0.0 => maze.io/x/ttyrec latest
)