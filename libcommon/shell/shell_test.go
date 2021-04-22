package shell

import (
	"testing"
	"time"
	"strings"
)

func TestShell_Exec(t *testing.T) {
	shell := &Shell{
		Cmd: "lsof -i:3306 -sTCP:LISTEN | wc -l",
		Args: map[string]string{
			"hello": "world",
		},
		Timeout: time.Second,
	}

	stdout, stderr, err := shell.Exec()
	if err != nil {
		t.Error("error:", err, "stderr:", string(stderr))
		return
	}

	t.Log(len(stdout), string(strings.TrimSpace(string(stdout))), string(stderr))
}
