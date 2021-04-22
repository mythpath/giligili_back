package shell

import (
	"time"
	"os/exec"
	"fmt"
	"bytes"
	"syscall"
	"errors"
	"io"
	"math"
)

var (
	ErrExecShellTimeout = errors.New("exec shell timeout")
)

type Shell struct {
	Cmd 		string
	Args 		map[string]string
	Timeout 	time.Duration

	shutdownCh 	chan struct{}
}

func NewShell(cmd string, args map[string]string, timeout time.Duration) *Shell {
	return &Shell{
		Cmd: cmd,
		Args: args,
		Timeout: timeout,
		shutdownCh: make(chan struct{}, 1),
	}
}

func (s *Shell) Release() {
	close(s.shutdownCh)
}

func (s *Shell) Kill() {
	s.shutdownCh <- struct{}{}
}

func (s *Shell) String() string {
	return fmt.Sprintf("{ cmd: %s, timeout: %v }", s.Cmd, s.Timeout)
}

func (s *Shell) Exec() ([]byte, []byte, error) {
	var (
		shell = s.command()
		errCh = make(chan error, 1)
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	shell.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	shell.Stderr = &stderr
	shell.Stdout = &stdout

	if err := shell.Start(); err != nil {
		return nil, nil, err
	}

	// wait shell exit
	go func() {
		errCh <- shell.Wait()
		close(errCh)
	}()

	timeout := s.Timeout
	if timeout == 0 {
		timeout = math.MaxInt64
	}
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		// kill bin/sh and all child process
		syscall.Kill(-shell.Process.Pid, syscall.SIGKILL)

		return nil, nil, ErrExecShellTimeout
	case err := <-errCh:
		return stdout.Bytes(), stderr.Bytes(), err
	case <-s.shutdownCh:
		// kill bin/sh and all child process
		syscall.Kill(-shell.Process.Pid, syscall.SIGKILL)

		return stdout.Bytes(), stderr.Bytes(), nil
	}
}

func (s *Shell) ExecInProgress(stdout, stderr io.Writer) error {
	var (
		shell = s.command()
		errCh = make(chan error, 1)
	)

	shell.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	shell.Stderr = stderr
	shell.Stdout = stdout

	if err := shell.Start(); err != nil {
		return err
	}

	// wait shell exit
	go func() {
		errCh <- shell.Wait()
		close(errCh)
	}()

	timeout := s.Timeout
	if timeout == 0 {
		timeout = math.MaxInt64
	}
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		// kill bin/sh and all child process
		syscall.Kill(-shell.Process.Pid, syscall.SIGKILL)

		return ErrExecShellTimeout
	case err := <-errCh:
		return err
	case <-s.shutdownCh:
		// kill bin/sh and all child process
		syscall.Kill(-shell.Process.Pid, syscall.SIGKILL)

		return nil
	}

	return nil
}

func (s *Shell) command() *exec.Cmd {
	shell := fmt.Sprintf("cd ~ && %s", s.Cmd)
	if s.Args != nil {
		var input bytes.Buffer

		for key, value := range s.Args {
			input.WriteString(fmt.Sprintf(" %s=%s", key, value))
		}

		shell = fmt.Sprintf("export %s && cd ~ && %s", input.String(), s.Cmd)
	}

	return exec.Command("/bin/sh", "-cx", shell)
}
