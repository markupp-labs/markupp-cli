package cli

import (
	"context"
	"io"
	"strings"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// noteCall is one call the commands made to the notes service.
type noteCall struct {
	method string
	id     string
	vault  string
	slug   string
	text   string
	page   client.Page
	create schema.NoteCreateRequest
	update schema.NoteUpdateRequest
}

// fakeNotes is the notes service of a test: it records each call and answers with the
// body or the error set for the method.
type fakeNotes struct {
	replies map[string]string
	errs    map[string]error
	calls   []noteCall
}

func newFakeNotes(replies map[string]string) *fakeNotes {
	return &fakeNotes{replies: replies, errs: map[string]error{}}
}

func (f *fakeNotes) answer(call noteCall) (io.ReadCloser, error) {
	f.calls = append(f.calls, call)
	if err := f.errs[call.method]; err != nil {
		return nil, err
	}
	return io.NopCloser(strings.NewReader(f.replies[call.method])), nil
}

func (f *fakeNotes) ListNotes(_ context.Context, vault string, page client.Page) (io.ReadCloser, error) {
	return f.answer(noteCall{method: "list", vault: vault, page: page})
}

func (f *fakeNotes) GetNote(_ context.Context, id string) (io.ReadCloser, error) {
	return f.answer(noteCall{method: "get", id: id})
}

func (f *fakeNotes) SearchNotes(_ context.Context, text, vault string, page client.Page) (io.ReadCloser, error) {
	return f.answer(noteCall{method: "search", text: text, vault: vault, page: page})
}

func (f *fakeNotes) PublicNotes(_ context.Context, slug string, _ client.NoteFilter, page client.Page) (io.ReadCloser, error) {
	return f.answer(noteCall{method: "public list", slug: slug, page: page})
}

func (f *fakeNotes) PublicNote(_ context.Context, slug, id string) (io.ReadCloser, error) {
	return f.answer(noteCall{method: "public get", slug: slug, id: id})
}

func (f *fakeNotes) PublicSearch(_ context.Context, slug, text string, page client.Page) (io.ReadCloser, error) {
	return f.answer(noteCall{method: "public search", slug: slug, text: text, page: page})
}

func (f *fakeNotes) CreateNote(_ context.Context, vault string, note schema.NoteCreateRequest) (io.ReadCloser, error) {
	return f.answer(noteCall{method: "create", vault: vault, create: note})
}

func (f *fakeNotes) UpdateNote(_ context.Context, id string, update schema.NoteUpdateRequest) (io.ReadCloser, error) {
	return f.answer(noteCall{method: "update", id: id, update: update})
}

func (f *fakeNotes) DeleteNote(_ context.Context, id string) error {
	_, err := f.answer(noteCall{method: "delete", id: id})
	return err
}

// notesMu returns a mu connected to notes, with a server URL configured. notes also
// answers the reads of a public vault.
func notesMu(notes *fakeNotes) *testMu {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	mu.server.services = Services{Notes: notes, Public: notes}
	return mu
}

const (
	noteListBody = `[{"id":"3b241101-e2bb-4255-8caf-4136c566a962","title":"meeting","path":"notas/meeting.md",` +
		`"updated_at":"2026-05-14T12:00:00.123456Z","content_sha256":"ab","degree":2,"pagerank":0.5,"community":1},` +
		`{"id":"n2","title":"b","path":"b.md","updated_at":"2026-05-15T08:30:00Z","degree":null,"pagerank":null,"community":null}]`
	noteBody = `{"id":"n1","vault_id":"v1","path":"notas/meeting.md","content":"# Meeting\n- item \u001b[31mred\n",` +
		`"content_sha256":"0664077f33cc3ebbaa4bbdacac0eb70e740983080f01dce29929e73b7785a7ad",` +
		`"created_at":"2026-05-14T12:00:00Z","updated_at":"2026-05-15T12:00:00.5Z","degree":15,"pagerank":0.0123,"community":null}`
)
