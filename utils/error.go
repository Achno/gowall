package utils

import (
	"fmt"
	"os"
)

// Prints the error in red and exits
func HandleError(err error, msg ...string) {
	if err != nil {

		switch {

		case len(msg) > 0:
			fmt.Printf("%s %s: %s %s\n", redColor, msg[0], err, resetColor)

		default:
			fmt.Printf("%s %s %s\n", redColor, err, resetColor)

		}

		os.Exit(1)
	}
}

// Formats a slice of errors to a single error, each seperated by a new line
func FormatErrors(errs []error) string {
	var result string
	for _, err := range errs {
		result += err.Error() + "\n"
	}
	return result
}
