package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LastRun is the input and output of the last single image command, used by `gowall compare` without args.
type LastRun struct {
	Original string `json:"original"`
	Edited   string `json:"edited"`
}

func SaveLastRun(run LastRun) error {
	if err := os.MkdirAll(filepath.Dir(LastRunFilePath), 0755); err != nil {
		return err
	}
	b, err := json.Marshal(run)
	if err != nil {
		return err
	}
	return os.WriteFile(LastRunFilePath, b, 0644)
}

func ReadLastRun() (LastRun, error) {
	b, err := os.ReadFile(LastRunFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return LastRun{}, fmt.Errorf("there is no previous run to compare, run something like `gowall convert` first")
	}
	if err != nil {
		return LastRun{}, err
	}
	var run LastRun
	if err := json.Unmarshal(b, &run); err != nil {
		return LastRun{}, fmt.Errorf("while reading %s: %w", LastRunFilePath, err)
	}
	return run, nil
}
