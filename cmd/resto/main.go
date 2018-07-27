package main

import (
	"os"

	"github.com/docker/docker/pkg/term"
	"github.com/sirupsen/logrus"
)

func main() {
	// Set terminal emulation based on platform as required.
	_, _, stderr := term.StdStreams()
	logrus.SetOutput(stderr)

	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
