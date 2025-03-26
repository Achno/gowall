package utils

import (
	"github.com/Achno/gowall/internal/logger"
)

// Prints the error in red and exits
func HandleError(err error, msg ...string) {
	if err != nil {
		switch {

		case len(msg) > 0:
			logger.Fatalf("%s: %s", msg[0], err)
		default:
			logger.Fatalf(err.Error())
		}
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
