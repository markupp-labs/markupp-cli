package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
)

// headerHeight is the number of lines of the header, and of hotkeys per column.
const headerHeight = 6

// header shows the session on the left and the hotkeys of the view on the right.
type header struct {
	widget *tview.Flex
	info   *tview.TextView
	keys   *tview.TextView
}

func newHeader() *header {
	h := &header{
		info: tview.NewTextView().SetDynamicColors(true).SetWrap(false),
		keys: tview.NewTextView().SetDynamicColors(true).SetWrap(false),
	}
	h.widget = tview.NewFlex().AddItem(h.info, 48, 0, false).AddItem(h.keys, 0, 1, false)
	return h
}

// headerState is everything the header draws.
type headerState struct {
	session Session
	tenant  string
	account string
	vault   string
	refresh string
	now     time.Time
	keys    []Hotkey
}

func (s *shell) headerState(top View) headerState {
	return headerState{
		session: s.session, tenant: s.scopeLabel(), account: s.account,
		vault: s.vault, refresh: s.refreshText(), now: s.cfg.now(), keys: top.Hotkeys(),
	}
}

func (h *header) render(state headerState) {
	h.info.SetText(infoText(state))
	h.keys.SetText(hotkeyColumns(state.keys, headerHeight))
}

// infoText is the left side of the header, one setting per line.
func infoText(state headerState) string {
	lines := [][2]string{
		{"Server", state.session.Server},
		{"Tenant", state.tenant},
		{"Account", state.account},
		{"Vault", state.vault},
		{"Token", tokenText(state.session, state.now)},
		{"Refresh", state.refresh},
	}
	var out strings.Builder
	for _, line := range lines {
		fmt.Fprintf(&out, "[yellow]%-8s[-] %s\n", line[0]+":", plain(orDash(line[1])))
	}
	return out.String()
}

func orDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

// tokenText says where the token comes from and when a stored login expires.
func tokenText(session Session, now time.Time) string {
	if session.Credential != "login" || session.ExpiresAt == nil {
		return session.Credential
	}
	left := session.ExpiresAt.Sub(now).Round(time.Minute)
	if left <= 0 {
		return "login, expired, renewed on the next request"
	}
	return "login, expires in " + left.String()
}

// hotkeyColumns lays the hotkeys out in columns of rows lines.
func hotkeyColumns(keys []Hotkey, rows int) string {
	lines := make([]string, rows)
	for i, key := range majorKeys(keys) {
		cell := fmt.Sprintf("[blue]%-9s[-]%-20s", tview.Escape(key.Name()), tview.Escape(key.Label))
		lines[i%rows] += cell
	}
	return strings.Join(lines, "\n")
}

// majorKeys leaves out the keys that show only in ? help.
func majorKeys(keys []Hotkey) []Hotkey {
	var major []Hotkey
	for _, key := range keys {
		if !key.minor {
			major = append(major, key)
		}
	}
	return major
}

// refreshText describes the auto refresh for the header.
func (s *shell) refreshText() string {
	switch {
	case s.cfg.Refresh <= 0:
		return "off"
	case s.paused:
		return "paused"
	default:
		return "every " + s.cfg.Refresh.String()
	}
}
