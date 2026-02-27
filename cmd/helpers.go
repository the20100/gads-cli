package cmd

import (
	"os/exec"
	"runtime"
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
