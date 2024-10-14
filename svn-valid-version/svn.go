package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

type SVN struct {
	Path string
}

func (s *SVN) Init(path string) {
	if path == "" {
		path = "svn"
	}
	s.Path = path
}

func (s *SVN) Blame(file string) (string, error) {
	cmd := exec.Command(s.Path, "blame", "--xml", file)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run svn blame: %w", err)
	}

	return out.String(), nil
}
