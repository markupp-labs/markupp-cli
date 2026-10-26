package tui

import (
	"io"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Config is everything mu tui takes from the command that starts it.
type Config struct {
	// Connect builds the ports of a scope; it is required.
	Connect Connector
	// Session describes the server and the login for the header. It is called at
	// start and on every auto refresh, so the expiry follows a renewed login.
	Session func() Session
	// Scope is the scope the interface starts in.
	Scope Scope
	// Vault is the vault :notes opens without an argument.
	Vault string
	// Refresh reloads the current listing at this interval; zero turns it off.
	Refresh time.Duration
	// Editor is the command line of $EDITOR; empty means the built in text area.
	Editor string
	// Terminal is where the editor runs and where y writes the OSC 52 sequence.
	Terminal Terminal
	// TempDir holds the temporary file of the editor; empty means the system default.
	TempDir string
	// Clock is the current time; nil means time.Now.
	Clock func() time.Time
	// Screen draws the interface; nil means the terminal. Tests pass a simulation.
	Screen tcell.Screen
	// FlashFor is how long a flash message stays; zero means flashDuration.
	FlashFor time.Duration
}

// Session is what the header shows about the server and the login. Tokens never
// reach the interface.
type Session struct {
	Server string
	// Credential is where the token comes from: env, login or none.
	Credential string
	// ExpiresAt is the expiry of a stored login, nil otherwise.
	ExpiresAt *time.Time
}

// Terminal is the terminal behind the screen, for the programs mu tui hands it to.
type Terminal struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

const flashDuration = 4 * time.Second

func (c Config) now() time.Time { return *new(time.Time) }

func (c Config) flashFor() time.Duration { return *new(time.Duration) }

func (c Config) session() Session { return *new(Session) }
