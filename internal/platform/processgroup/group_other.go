//go:build !unix

package processgroup

import (
	"os"
	"os/exec"
)

func startInOwnGroup(*exec.Cmd) {}

func interruptGroup(process *os.Process) error {
	return process.Signal(os.Interrupt)
}

func killGroup(process *os.Process) {
	_ = process.Kill()
}
