package client

import (
	"context"
	"io"
	"net/http"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

func publicationPath(id string) string { return endpoint("vaults", id, "publication") }

// GetPublication returns the publication of the vault id. A vault that is not
// published answers 404 with the code not_published.
func (c *Client) GetPublication(ctx context.Context, id string) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: publicationPath(id)})
}

// PublishVault publishes the vault id under slug, or changes the slug and title of a
// vault already published. An empty title is not sent.
func (c *Client) PublishVault(ctx context.Context, id, slug, title string) (io.ReadCloser, error) {
	r := request{method: http.MethodPut, path: publicationPath(id)}
	return openJSON(ctx, c, r, schema.PublicationRequest{Slug: slug, Title: title})
}

// UnpublishVault takes the vault id out of the public reads and frees its slug.
func (c *Client) UnpublishVault(ctx context.Context, id string) error {
	return c.exec(ctx, request{method: http.MethodDelete, path: publicationPath(id)})
}
