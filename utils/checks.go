package utils

import (
	"fmt"
	"os"
	"strings"
)

// Checks if the terminal using gowall is the kitty terminal emulator
func IsKittyTerminalRunning() bool {

	terminal := os.Getenv("TERM")
	kittyInstanceId := os.Getenv("KITTY_WINDOW_ID")

	return strings.Contains(terminal, "kitty") || kittyInstanceId != ""
}

// Checks if the terminal running is Konsole and has
func IsKonsoleTerminalRunning() bool {

	terminal := os.Getenv("TERM")

	if terminal == "xterm-256color" && os.Getenv("KONSOLE_VERSION") != "" {
		return true
	}
	return false
}

// Checks if the terminal running is Ghostty and has
func IsGhosttyTerminalRunning() bool {

	terminal := os.Getenv("TERM")

	if strings.Contains(terminal, "ghostty") {
		return true
	}

	return false
}

func IsWeztermTerminalRunning() bool {

	terminal := os.Getenv("TERM")

	if terminal == "xterm-256color" && os.Getenv("TERM_PROGRAM") == "WezTerm" {
		return true
	}
	return false
}

func Confirm(msg string) bool {

	var input string

	fmt.Printf("%s (y/n): ", msg)
	fmt.Scanln(&input)

	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y"
}
