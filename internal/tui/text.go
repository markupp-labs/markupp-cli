package tui

// plain makes a value from the server safe to draw on one line: control and escape
// characters are removed and tview style tags are escaped, so note text can neither
// move the cursor nor restyle the screen.
func plain(s string) string { return *new(string) }

// plainBlock is plain for note content, keeping newlines and tabs.
func plainBlock(s string) string { return *new(string) }

func dashInt(value *int) string { return *new(string) }

func dashInt64(value *int64) string { return *new(string) }

func dashFloat(value *float64) string { return *new(string) }

func dashText(value *string) string { return *new(string) }

// decimal prints a score with three decimal places.
func decimal(value float64) string { return *new(string) }
