package cli

import (
	"context"
	"io"
	"strings"

	"github.com/markupp-labs/markupp-cli/internal/client"
)

// portCall is one call a fake port received: the method and its arguments.
type portCall struct {
	method   string
	subject  string
	linkType string
	page     client.Page
	graph    client.GraphFilter
}

// recordingPort records each call and answers with body, or with err when it is set.
type recordingPort struct {
	body  string
	err   error
	calls []portCall
}

func (p *recordingPort) answer(call portCall) (io.ReadCloser, error) {
	p.calls = append(p.calls, call)
	if p.err != nil {
		return nil, p.err
	}
	return io.NopCloser(strings.NewReader(p.body)), nil
}

// fakeNoteLinks is a NoteLinkReader over a recordingPort.
type fakeNoteLinks struct{ recordingPort }

func (f *fakeNoteLinks) NoteLinks(_ context.Context, id, linkType string, page client.Page) (io.ReadCloser, error) {
	return f.answer(portCall{method: "NoteLinks", subject: id, linkType: linkType, page: page})
}

func (f *fakeNoteLinks) NoteBacklinks(_ context.Context, id, linkType string, page client.Page) (io.ReadCloser, error) {
	return f.answer(portCall{method: "NoteBacklinks", subject: id, linkType: linkType, page: page})
}

func (f *fakeNoteLinks) NoteMentions(_ context.Context, id string, page client.Page) (io.ReadCloser, error) {
	return f.answer(portCall{method: "NoteMentions", subject: id, page: page})
}

func (f *fakeNoteLinks) NoteBackmentions(_ context.Context, id string, page client.Page) (io.ReadCloser, error) {
	return f.answer(portCall{method: "NoteBackmentions", subject: id, page: page})
}

func (f *fakeNoteLinks) NoteSuggestions(_ context.Context, id string, page client.Page) (io.ReadCloser, error) {
	return f.answer(portCall{method: "NoteSuggestions", subject: id, page: page})
}

// linkedMu is a mu with a server URL whose note link port answers body.
func linkedMu(body string) (*testMu, *fakeNoteLinks) {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	links := &fakeNoteLinks{recordingPort{body: body}}
	mu.server.services = Services{NoteLinks: links}
	return mu, links
}
