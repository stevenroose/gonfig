// +build windows

package gonfig

import (
	"golang.org/x/sys/windows"
)

// getTerminalWidth returns the width of the current terminal.
func getTerminalWidth() int {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(0), &info); err != nil {
		return 0
	}
	return int(info.Size.X)
}
