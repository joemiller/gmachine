package gcp

import (
	"io"
	"os"
	"os/exec"
	"syscall"
)

// run runs a command and connects the command's stdout and stderr to the provided
// io.Writer's. Typically this will be os.Stdout and os.Stderr. An optional io.reader
// can be passed as stdin to the command.
var run = func(stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	exe := exec.Command(args[0], args[1:]...)
	exe.Stdout = stdout
	exe.Stderr = stderr
	exe.Stdin = stdin
	env := os.Environ()
	env = append(env, "PYTHONUNBUFFERED=1")
	exe.Env = env
	return exe.Run()
}

// execve replaces the current process with a new process.
var execve = func(args []string) error {
	path, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	// TODO: refactor? I think we're expected to use the unix pkg from x/sys instead of syscall. https://godoc.org/golang.org/x/sys/unix#Exec
	return syscall.Exec(path, args, os.Environ())
}

var output = func(args ...string) ([]byte, error) {
	return exec.Command(args[0], args[1:]...).CombinedOutput()
}
