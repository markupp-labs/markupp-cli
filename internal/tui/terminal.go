package tui

import (
	"fmt"
	"os"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// checkTerminal refuses a stdin or stdout redirected to a file or a pipe: the screen
// is drawn on the controlling terminal, but y copies through stdout and the editor
// runs on stdin and stdout, so a redirect would get the OSC 52 sequence and leave
// the editor without a terminal. A stream that is not a file, as in tests, passes.
func checkTerminal(term Terminal) error {
	in, _ := term.In.(*os.File)
	out, _ := term.Out.(*os.File)
	for name, file := range map[string]*os.File{"stdin": in, "stdout": out} {
		if file != nil && !isCharDevice(file) {
			message := fmt.Sprintf("%s is redirected to %q, expected the terminal mu tui runs in", name, file.Name())
			return failure.New(failure.Usage, "mu tui", message)
		}
	}
	return nil
}

func isCharDevice(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
