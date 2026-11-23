//go:build !unix

package pythondownloader

import (
	"os"
	"os/exec"
)

func startInOwnProcessGroup(*exec.Cmd) {}

func interruptProcessGroup(process *os.Process) error {
	return process.Signal(os.Interrupt)
}

func killProcessGroup(process *os.Process) {
	_ = process.Kill()
}
