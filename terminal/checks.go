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
	return terminal == "xterm-256color" && os.Getenv("KONSOLE_VERSION") != ""
}

// Checks if the terminal running is Ghostty
func IsGhosttyTerminalRunning() bool {
	terminal := os.Getenv("TERM")
	return terminal == "xterm-ghostty" && os.Getenv("TERM_PROGRAM") == "ghostty"
}

// Checks if the terminal running is Wezterm
func IsWeztermTerminalRunning() bool {
	terminal := os.Getenv("TERM")
	return terminal == "xterm-256color" && os.Getenv("TERM_PROGRAM") == "WezTerm"
}

// Checks if the user has the kitten binary installed, so the kitten icat image utility can be used
func HasIcat() bool {
	_, err := exec.LookPath("kitten")
	return err == nil
}

// Checks if the user has chafa in his $PATH
func HasChafa() bool {
	_, err := exec.LookPath("chafa")
	return err == nil
}
