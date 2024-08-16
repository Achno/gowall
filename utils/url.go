package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

// opens a URL in your default browser of your operating system
func OpenURL(url string) error {

	var cmd *exec.Cmd

	switch runtime.GOOS {

	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)

	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()

}
