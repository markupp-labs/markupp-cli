package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// stdinFile is the --file value that reads the content from stdin.
const stdinFile = "-"

// addFileFlag adds --file, where a write command takes the note content from. There is
// no implicit stdin: reading it takes --file -.
func addFileFlag(cmd *cobra.Command, file *string) {
	cmd.Flags().StringVar(file, "file", "", "file holding the note content, or - for stdin (required)")
}

// readContent reads the note content named by --file: a path, or - for stdin.
func (a *App) readContent(file string) (string, error) {
	switch file {
	case "":
		return "", failure.New(failure.Usage, "--file", "missing, expected a path or - for stdin")
	case stdinFile:
		return readStdin(a.deps.Stdin)
	}
	raw, err := a.readFile()(file)
	if err != nil {
		return "", failure.Wrap(failure.Usage, fmt.Sprintf("--file %q", file), err)
	}
	return string(raw), nil
}

func readStdin(stdin io.Reader) (string, error) {
	if stdin == nil {
		return "", failure.New(failure.Usage, "--file -", "no stdin to read, expected content piped in")
	}
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return "", failure.Wrap(failure.Usage, "--file - (stdin)", err)
	}
	return string(raw), nil
}

func (a *App) readFile() func(string) ([]byte, error) {
	if a.deps.ReadFile != nil {
		return a.deps.ReadFile
	}
	return os.ReadFile
}
