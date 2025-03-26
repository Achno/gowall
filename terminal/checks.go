package terminal

import (
	"os"
	"os/exec"
	"strings"
)

// Checks if the terminal using gowall is the kitty terminal emulator
func IsKittyTerminalRunning() bool {

	terminal := os.Getenv("TERM")
	kittyInstanceId := os.Getenv("KITTY_WINDOW_ID")

	return strings.Contains(terminal, "kitty") || kittyInstanceId != ""
}

// Checks if the terminal running is Konsole
func IsKonsoleTerminalRunning() bool {

	terminal := os.Getenv("TERM")

	if terminal == "xterm-256color" && os.Getenv("KONSOLE_VERSION") != "" {
		return true
	}
	return false
}

// Checks if the terminal running is Ghostty
func IsGhosttyTerminalRunning() bool {

	terminal := os.Getenv("TERM")

	if terminal == "xterm-ghostty" && os.Getenv("TERM_PROGRAM") == "ghostty" {
		return true
	}
	return false
}

// Checks if the terminal running is Wezterm
func IsWeztermTerminalRunning() bool {

	terminal := os.Getenv("TERM")

	if terminal == "xterm-256color" && os.Getenv("TERM_PROGRAM") == "WezTerm" {
		return true
	}
	return false
}

// Checks if the user has the kitten binary installed, so the kitten icat image utility can be used
func HasIcat() bool {
	path, err := exec.LookPath("kitten")
	if err != nil {
		return false
	}

	return path != ""
}

// Checks if the user has chafa in his $PATH
func HasChafa() bool {
	path, err := exec.LookPath("chafa")
	if err != nil {
		return false
	}

	return path != ""
}
