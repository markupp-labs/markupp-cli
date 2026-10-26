package output

import "strings"

// The search snippet markers of the API, and how the text output shows them.
const (
	snippetStart     = '\u0002'
	snippetEnd       = '\u0003'
	snippetStartText = "[["
	snippetEndText   = "]]"
)

type cleaning struct {
	keepLayout bool
	markers    bool
}

// SanitizeLine removes control and escape characters and folds newlines and tabs
// into spaces, so a value from the server stays on its line.
func SanitizeLine(s string) string { return clean(s, cleaning{}) }

// SanitizeBlock removes control and escape characters, keeping newlines and tabs.
// Carriage returns are dropped, so CRLF text prints as LF.
func SanitizeBlock(s string) string { return clean(s, cleaning{keepLayout: true}) }

// Snippet renders a search snippet on one line, with the match markers U+0002 and
// U+0003 shown as "[[" and "]]". The text is sanitized first, so a marker can only
// come from the server's highlighting.
func Snippet(s string) string { return clean(s, cleaning{markers: true}) }

func clean(s string, how cleaning) string {
	var out strings.Builder
	out.Grow(len(s))
	for _, r := range s {
		writeClean(&out, r, how)
	}
	return out.String()
}

func writeClean(out *strings.Builder, r rune, how cleaning) {
	switch {
	case how.markers && r == snippetStart:
		out.WriteString(snippetStartText)
	case how.markers && r == snippetEnd:
		out.WriteString(snippetEndText)
	case r == '\n' || r == '\t':
		out.WriteRune(layoutRune(r, how.keepLayout))
	case isUnsafe(r):
		return
	default:
		out.WriteRune(r)
	}
}

func layoutRune(r rune, keepLayout bool) rune {
	if keepLayout {
		return r
	}
	return ' '
}

// isUnsafe reports C0 and C1 controls, DEL, and the bidirectional overrides and
// isolates that can make text display differently from what it is.
func isUnsafe(r rune) bool {
	switch {
	case r < 0x20, r == 0x7f, r >= 0x80 && r <= 0x9f:
		return true
	case r >= 0x202a && r <= 0x202e, r >= 0x2066 && r <= 0x2069:
		return true
	case r == 0x200e, r == 0x200f, r == 0x061c:
		return true
	}
	return false
}
