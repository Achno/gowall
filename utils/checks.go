package utils

import (
	"fmt"
	"strings"

	"github.com/Achno/gowall/internal/logger"
)

func Confirm(msg string) bool {
	var input string

	logger.Printf("%s (y/n): ", msg)
	fmt.Scanln(&input)

	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y"
}
