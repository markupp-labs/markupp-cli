package output

import "testing"

func TestSanitizeLine(t *testing.T) {
	cases := map[string]string{
		"plain":                         "plain",
		"two\nlines\tand tab":           "two lines and tab",
		"red \x1b[31mtext\x1b[0m":       "red [31mtext[0m",
		"bell\x07 cr\r del\x7f":         "bell cr del",
		"c1 \u009bcsi":                  "c1 csi",
		"bidi \u202eevil\u2066x":        "bidi evilx",
		"acentuação e emoji \U0001F600": "acentuação e emoji \U0001F600",
	}
	for in, want := range cases {
		if got := SanitizeLine(in); got != want {
			t.Errorf("SanitizeLine(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeBlockKeepsLayout(t *testing.T) {
	in := "# Title\r\n\n\tindented \x1b]0;pwned\x07line\n"
	want := "# Title\n\n\tindented ]0;pwnedline\n"
	if got := SanitizeBlock(in); got != want {
		t.Errorf("SanitizeBlock(%q) = %q, want %q", in, got, want)
	}
}

func TestSnippetReplacesMarkersAfterSanitizing(t *testing.T) {
	in := "about \x02kubernetes\x03 in\nthe \x1bcluster"
	want := "about [[kubernetes]] in the cluster"
	if got := Snippet(in); got != want {
		t.Errorf("Snippet(%q) = %q, want %q", in, got, want)
	}
}
