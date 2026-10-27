package tui

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Screen size of the tests: wide enough for the header and a table.
const (
	screenWidth  = 140
	screenHeight = 30
	waitTimeout  = 5 * time.Second
)

// fixedNow is the clock of every test.
var fixedNow = time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)

// syncBuffer is a terminal output the interface and the test both touch.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// harness is one running mu tui on a simulation screen.
type harness struct {
	t        *testing.T
	screen   tcell.SimulationScreen
	shell    *shell
	terminal *syncBuffer
	done     chan error
	cancel   context.CancelFunc
	stopped  sync.Once
}

// testConfig is a Config on the fake ports, with a fixed clock and a long flash.
func testConfig(connector *fakeConnector) Config {
	expires := fixedNow.Add(42 * time.Minute)
	return Config{
		Connect: connector.connect,
		Session: func() Session {
			return Session{Server: "https://markupp.example.com", Credential: "login", ExpiresAt: &expires}
		},
		Clock:    func() time.Time { return fixedNow },
		FlashFor: time.Minute,
	}
}

// startTUI runs mu tui and waits for its first frame at the test size.
func startTUI(t *testing.T, cfg Config, configure ...func(*shell)) *harness {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	terminal := &syncBuffer{}
	cfg.Screen, cfg.Terminal.Out = screen, terminal
	ctx, cancel := context.WithCancel(context.Background())
	s, err := newShell(ctx, cfg)
	if err != nil {
		cancel()
		t.Fatalf("newShell: %v", err)
	}
	for _, apply := range configure {
		apply(s)
	}
	h := &harness{t: t, screen: screen, shell: s, terminal: terminal, done: make(chan error, 1), cancel: cancel}
	h.launch()
	return h
}

// launch runs the shell, then resizes the screen to the test size.
func (h *harness) launch() {
	go func() { h.done <- h.shell.run() }()
	h.t.Cleanup(h.stop)
	h.waitFor("Server:")
	h.screen.SetSize(screenWidth, screenHeight)
	if err := h.screen.PostEvent(tcell.NewEventResize(screenWidth, screenHeight)); err != nil {
		h.t.Fatalf("resizing the screen: %v", err)
	}
	h.waitResized()
}

// waitResized waits for a frame drawn at the full width of the screen.
func (h *harness) waitResized() {
	h.t.Helper()
	eventually(h.t, "a frame at the test size", func() bool {
		for _, line := range strings.Split(h.text(), "\n") {
			if len([]rune(line)) == screenWidth {
				return true
			}
		}
		return false
	})
}

// stop ends mu tui once, however many times it is called.
func (h *harness) stop() {
	h.stopped.Do(h.stopNow)
}

func (h *harness) stopNow() {
	h.cancel()
	select {
	case <-h.done:
	case <-time.After(waitTimeout):
		h.t.Errorf("mu tui did not stop within %s", waitTimeout)
	}
}

// text is the screen as lines of text, trailing spaces removed. It reads the screen
// on the event loop, which is the goroutine that draws it.
func (h *harness) text() string {
	screen := make(chan string, 1)
	h.shell.app.QueueUpdate(func() { screen <- h.contents() })
	select {
	case text := <-screen:
		return text
	case <-time.After(waitTimeout):
		return "(the event loop is not running)\n"
	}
}

func (h *harness) contents() string {
	cells, width, _ := h.screen.GetContents()
	var lines []string
	for row := 0; row*width < len(cells); row++ {
		lines = append(lines, strings.TrimRight(rowText(cells[row*width:(row+1)*width]), " "))
	}
	return strings.Join(lines, "\n") + "\n"
}

func rowText(cells []tcell.SimCell) string {
	var line strings.Builder
	for _, cell := range cells {
		if len(cell.Runes) == 0 {
			line.WriteRune(' ')
			continue
		}
		line.WriteString(string(cell.Runes))
	}
	return line.String()
}

// waitFor waits until the screen shows every one of texts.
func (h *harness) waitFor(texts ...string) string {
	h.t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		screen := h.text()
		if containsAll(screen, texts) {
			return screen
		}
		time.Sleep(5 * time.Millisecond)
	}
	h.t.Fatalf("the screen never showed %q:\n%s", texts, h.text())
	return ""
}

// waitGone waits until the screen no longer shows text.
func (h *harness) waitGone(text string) {
	h.t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		if !strings.Contains(h.text(), text) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	h.t.Fatalf("the screen kept showing %q:\n%s", text, h.text())
}

func containsAll(screen string, texts []string) bool {
	for _, text := range texts {
		if !strings.Contains(screen, text) {
			return false
		}
	}
	return true
}

// typeText types each rune as a key press.
func (h *harness) typeText(text string) {
	for _, r := range text {
		h.screen.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	}
}

// press presses a control key.
func (h *harness) press(key tcell.Key) {
	h.screen.InjectKey(key, 0, tcell.ModNone)
}

// command types a : command and presses Enter.
func (h *harness) command(line string) {
	h.typeText(":" + line)
	h.press(tcell.KeyEnter)
}

// onUI runs f on the event loop and waits for it, to read the state of the shell.
func (h *harness) onUI(f func()) {
	h.t.Helper()
	ran := make(chan struct{})
	h.shell.app.QueueUpdate(func() { f(); close(ran) })
	select {
	case <-ran:
	case <-time.After(waitTimeout):
		h.t.Fatalf("the event loop did not run the update within %s", waitTimeout)
	}
}

// eventually waits until check holds, for state the fakes record off the event loop.
func eventually(t *testing.T, what string, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("%s never happened", what)
}
