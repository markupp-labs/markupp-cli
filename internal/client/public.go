package client

import (
	"context"
	"io"
)

// PublicNotes lists the notes of the public vault slug, without content.
func (c *Client) PublicNotes(context.Context, string, NoteFilter, Page) (io.ReadCloser, error) {
	return nil, errPublicationPending
}

// PublicNote returns the note id of the public vault slug, with its content.
func (c *Client) PublicNote(context.Context, string, string) (io.ReadCloser, error) {
	return nil, errPublicationPending
}

// PublicSearch searches the public vault slug.
func (c *Client) PublicSearch(context.Context, string, string, Page) (io.ReadCloser, error) {
	return nil, errPublicationPending
}

// PublicGraph returns the graph of the public vault slug.
func (c *Client) PublicGraph(context.Context, string, GraphFilter) (io.ReadCloser, error) {
	return nil, errPublicationPending
}

// PublicLinks lists the links of the public vault slug.
func (c *Client) PublicLinks(context.Context, string, LinkFilter, Page) (io.ReadCloser, error) {
	return nil, errPublicationPending
}
