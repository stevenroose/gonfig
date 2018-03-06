// +build !windows

package gonfig

import (
	"golang.org/x/sys/unix"
)

// getTerminalWidth returns the width of the current terminal.
func getTerminalWidth() int {
	ws, err := unix.IoctlGetWinsize(0, unix.TIOCGWINSZ)
	if err != nil {
		return 0
	}
	return int(ws.Col)
}
