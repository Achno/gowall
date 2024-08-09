package utils

import (
	"os"
	"strings"
)

// Checks if the terminal using gowall is the kitty terminal emulator
func IsKittyTerminalRunning() bool {

	terminal := os.Getenv("TERM")
	kittyInstanceId := os.Getenv("KITTY_WINDOW_ID")

	return strings.Contains(terminal, "kitty") || kittyInstanceId != ""
}
