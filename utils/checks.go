package utils

import (
	"fmt"
	"strings"
)

func Confirm(msg string) bool {

	var input string

	fmt.Printf("%s (y/n): ", msg)
	fmt.Scanln(&input)

	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y"
}
