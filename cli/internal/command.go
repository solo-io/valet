package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ExecuteCmd(filepath string, args ...string) (string, error) {
	return ExecuteCmdStdIn("", filepath, args...)
}

func ExecuteCmdStdIn(stdIn, filepath string, args ...string) (string, error) {
	cmd := exec.Command(filepath, args...)
	cmd.Stdin = strings.NewReader(stdIn)
	cmd.Env = os.Environ()
	for i, pair := range cmd.Env {
		if strings.HasPrefix(pair, "DEBUG") {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i+1:]...)
			break
		}
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%s (%v)", out, err)
	}
	return string(out), err
}
