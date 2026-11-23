// Package processgroup runs external tools so that cancelling reaches the children they start, like Ctrl-C would.
package processgroup

import (
	"os"
	"os/exec"
	"time"
)

// Interruptible makes a cancel interrupt the command's whole group, then lets os/exec kill the leader after grace.
func Interruptible(command *exec.Cmd, grace time.Duration) {
	startInOwnGroup(command)
	command.Cancel = func() error { return interruptGroup(command.Process) }
	command.WaitDelay = grace
}

// Kill ends what is left of a finished command's group: os/exec kills only the leader, and a crashed leader leaves its children behind.
func Kill(process *os.Process) {
	killGroup(process)
}
