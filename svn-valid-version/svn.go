package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

type SVN struct {
}

func (s *SVN) Blame(file string) (string, error) {
	cmd := exec.Command("svn", "blame", "--xml", file)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run svn blame: %w, %s", err, stderr.String())
	}

	return stdout.String(), nil
}
