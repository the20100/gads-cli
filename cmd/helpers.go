package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/the20100/gads-cli/internal/auth"
)

// openBrowser opens a URL in the default system browser.
func openBrowser(url string) {
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start() //nolint
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start() //nolint
	case "darwin":
		exec.Command("open", url).Start() //nolint
	}
}

// maskString masks all but the first 4 and last 4 chars of a string.
func maskString(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

// prompt reads a single line from stdin with a prompt message.
func prompt(message string) string {
	fmt.Print(message)
	var input string
	fmt.Scanln(&input)
	return input
}

// loadCreds loads credentials from the config file.
func loadCreds() (*auth.Credentials, error) {
	return auth.Load()
}
