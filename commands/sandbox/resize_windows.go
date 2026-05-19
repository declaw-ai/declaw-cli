//go:build windows

package sandbox

import (
	"os"
)

func sigwinch() os.Signal {
	return nil
}

func notifyResize(ch chan<- os.Signal) {}
