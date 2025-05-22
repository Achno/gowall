package utils

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/theckman/yacspin"
)

// global spinner
var (
	Spinner *yacspin.Spinner
)

// Creates a new global spinner
func NewSpinner(cfg yacspin.Config) {
	spinner, err := yacspin.New(cfg)
	if err != nil {
		log.Fatalf("Error: Could not create and load spinner: %v", err)
	}

	Spinner = spinner
}

// Sets the spinner to quiet mode by discarding all output to not interfere with Unix pipes or redirections
func SetSpinnerQuiet(quiet bool) {
	if quiet {
		cfg := SpinnerConfig()
		cfg.Writer = io.Discard
		NewSpinner(cfg)
	}
}

func SpinnerConfig() yacspin.Config {
	return yacspin.Config{
		Frequency:       200 * time.Millisecond,
		CharSet:         yacspin.CharSets[24],
		SuffixAutoColon: true,
		Writer:          os.Stdout,
		StopMessage:     "finished task successfully\n",
		StopCharacter:   "✓ ",
		StopColors:      []string{"fgGreen"},
		Colors:          []string{"fgBlue"},
		ShowCursor:      true,
	}
}
