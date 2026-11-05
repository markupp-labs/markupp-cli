package session

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// OpenBrowser asks the desktop to open target in the default browser and returns
// without waiting for it. It is best effort: login prints the URL anyway, for a machine
// without a browser.
func OpenBrowser(target string) error {
	name, args := browserCommand(runtime.GOOS, target)
	//nolint:gosec // G204: the command is xdg-open, open or cmd, and the URL is a single argument.
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %s to open the browser: %w", name, err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

// browserCommand picks the opener of goos. On Windows, start goes through cmd, which
// would split the query at "&" unless escaped; the empty argument is the window title
// start expects before the URL.
func browserCommand(goos, target string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{target}
	case "windows":
		return "cmd", []string{"/c", "start", "", strings.ReplaceAll(target, "&", "^&")}
	}
	return "xdg-open", []string{target}
}
