package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// open runs r and returns the response body unread, for the caller to stream or pass
// through. The caller closes it.
func (c *Client) open(ctx context.Context, r request) (io.ReadCloser, error) {
	resp, err := c.send(ctx, r)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// exec runs r and discards the body, for answers without content.
func (c *Client) exec(ctx context.Context, r request) error {
	resp, err := c.send(ctx, r)
	if err != nil {
		return err
	}
	discard(resp)
	return nil
}

var errNullArray = errors.New("null array")

// Elements streams the elements of a JSON array body one at a time, without reading
// the whole array, and closes body when the sequence ends or the consumer stops. A
// null body is an empty array. Anything else that is not an array is an error.
func Elements(body io.ReadCloser) iter.Seq2[json.RawMessage, error] {
	return func(yield func(json.RawMessage, error) bool) {
		defer func() { _ = body.Close() }()
		yieldElements(json.NewDecoder(body), yield)
	}
}

func yieldElements(decoder *json.Decoder, yield func(json.RawMessage, error) bool) {
	if err := openArray(decoder); err != nil {
		if !errors.Is(err, errNullArray) {
			yield(nil, err)
		}
		return
	}
	for decoder.More() {
		var element json.RawMessage
		if err := decoder.Decode(&element); err != nil {
			yield(nil, malformed(err))
			return
		}
		if !yield(element, nil) {
			return
		}
	}
}

func openArray(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return malformed(err)
	}
	if token == nil {
		return errNullArray
	}
	if delim, ok := token.(json.Delim); ok && delim == '[' {
		return nil
	}
	return malformed(fmt.Errorf("expected a JSON array, got %v", token))
}

func malformed(err error) error {
	return failure.Wrap(failure.Internal, "malformed server response", err)
}
