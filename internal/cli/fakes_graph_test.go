package cli

import (
	"context"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
)

// fakeGraph is a GraphReader over a recordingPort.
type fakeGraph struct{ recordingPort }

func (f *fakeGraph) VaultGraph(_ context.Context, vault string, filter client.GraphFilter) (io.ReadCloser, error) {
	return f.answer(portCall{method: "VaultGraph", subject: vault, graph: filter})
}

func (f *fakeGraph) PublicVaultGraph(_ context.Context, slug string, filter client.GraphFilter) (io.ReadCloser, error) {
	return f.answer(portCall{method: "PublicVaultGraph", subject: slug, graph: filter})
}

func (f *fakeGraph) NoteGraph(_ context.Context, id string, depth int, filter client.GraphFilter) (io.ReadCloser, error) {
	return f.answer(portCall{method: "NoteGraph", subject: id, depth: depth, graph: filter})
}

func (f *fakeGraph) VaultCommunities(_ context.Context, vault string, page client.Page) (io.ReadCloser, error) {
	return f.answer(portCall{method: "VaultCommunities", subject: vault, page: page})
}

func (f *fakeGraph) VaultCommunity(_ context.Context, vault string, cid int64) (io.ReadCloser, error) {
	return f.answer(portCall{method: "VaultCommunity", subject: vault, cid: cid})
}

func (f *fakeGraph) VaultLinks(_ context.Context, vault string, filter client.LinkFilter, page client.Page) (io.ReadCloser, error) {
	return f.answer(portCall{method: "VaultLinks", subject: vault, links: filter, page: page})
}

func (f *fakeGraph) PublicVaultLinks(_ context.Context, slug string, filter client.LinkFilter, page client.Page) (io.ReadCloser, error) {
	return f.answer(portCall{method: "PublicVaultLinks", subject: slug, links: filter, page: page})
}

// graphMu is a mu with a server URL whose graph port answers body.
func graphMu(body string) (*testMu, *fakeGraph) {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	graph := &fakeGraph{recordingPort{body: body}}
	mu.server.services = Services{Graph: graph}
	return mu, graph
}
