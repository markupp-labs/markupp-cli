package oauth

import (
	"context"
	"net/url"
)

// Revoke revokes refreshToken at the revocation endpoint (RFC 7009), which ends the
// authorization it belongs to. The server answers 200 even for a token it does not
// know, so a success says nothing about whether the token was valid.
func (c *Client) Revoke(ctx context.Context, meta Metadata, refreshToken string) error {
	form := url.Values{
		"token":           {refreshToken},
		"token_type_hint": {"refresh_token"},
		"client_id":       {c.cfg.ClientID},
	}
	resp, err := c.post(ctx, meta.RevocationEndpoint, form)
	if err != nil {
		return err
	}
	discard(resp)
	return nil
}
