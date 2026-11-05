package tui

import (
	"strconv"

	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/output"
)

// plain makes a value from the server safe to draw on one line: control and escape
// characters are removed and tview style tags are escaped, so note text can neither
// move the cursor nor restyle the screen.
func plain(s string) string { return tview.Escape(output.SanitizeLine(s)) }

// plainBlock is plain for note content, keeping newlines and tabs.
func plainBlock(s string) string { return tview.Escape(output.SanitizeBlock(s)) }

func dashInt(value *int) string {
	if value == nil {
		return "-"
	}
	return strconv.Itoa(*value)
}

func dashInt64(value *int64) string {
	if value == nil {
		return "-"
	}
	return strconv.FormatInt(*value, 10)
}

func dashFloat(value *float64) string {
	if value == nil {
		return "-"
	}
	return decimal(*value)
}

func dashText(value *string) string {
	if value == nil || *value == "" {
		return "-"
	}
	return *value
}

// decimal prints a score with three decimal places.
func decimal(value float64) string { return strconv.FormatFloat(value, 'f', 3, 64) }
