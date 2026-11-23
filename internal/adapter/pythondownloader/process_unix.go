//go:build unix

package pythondownloader

import (
	"os"
	"os/exec"
	"syscall"
)

func startInOwnProcessGroup(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// yt-dlp and aria2c are children of the downloader, so the interrupt goes to the whole group, like Ctrl-C would.
func interruptProcessGroup(process *os.Process) error {
	return syscall.Kill(-process.Pid, syscall.SIGINT)
}

// os/exec kills only the leader once the grace period ends, and a crashed leader leaves its children behind.
func killProcessGroup(process *os.Process) {
	_ = syscall.Kill(-process.Pid, syscall.SIGKILL)
}
