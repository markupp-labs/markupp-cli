package client

import (
	"context"
	"io"
)

// GraphFilter narrows a graph.
type GraphFilter struct {
	HideOrphans    bool
	HideUnresolved bool
	HideInferred   bool
	Community      *int64
}

func (c *Client) VaultGraph(context.Context, string, GraphFilter) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) PublicVaultGraph(context.Context, string, GraphFilter) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) NoteGraph(context.Context, string, int, GraphFilter) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) VaultCommunities(context.Context, string, Page) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

func (c *Client) VaultCommunity(context.Context, string, int64) (io.ReadCloser, error) {
	return nil, errNotImplemented
}
