//go:build !windows

package sandbox

import (
	"os"
	"os/signal"
	"syscall"
)

func sigwinch() os.Signal {
	return syscall.SIGWINCH
}

func notifyResize(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGWINCH)
}
