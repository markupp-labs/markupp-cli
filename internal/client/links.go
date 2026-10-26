package client

import (
	"context"
	"errors"
	"io"
)

// LinkFilter narrows a listing of links. An empty field is not sent.
type LinkFilter struct {
	Status string
	Type   string
}

var errNotImplemented = errors.New("not implemented")

func (c *Client) NoteLinks(context.Context, string, string, Page) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) NoteBacklinks(context.Context, string, string, Page) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) NoteMentions(context.Context, string, Page) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) NoteBackmentions(context.Context, string, Page) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) NoteSuggestions(context.Context, string, Page) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) VaultLinks(context.Context, string, LinkFilter, Page) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) PublicVaultLinks(context.Context, string, LinkFilter, Page) (io.ReadCloser, error) {
	return nil, errNotImplemented
}
