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

func Confirm(msg string) bool {

	var input string

	fmt.Printf("%s (y/n): ", msg)
	fmt.Scanln(&input)

	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y"
}
