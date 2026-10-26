package client

import (
	"context"
	"io"
	"testing"
)

// apiCall is one client method under test and the request it must send.
type apiCall struct {
	name string
	call func(ctx context.Context, c *Client) error
	want recordedRequest
}

// opened closes the body of a successful call, so a case only reports the error.
func opened(body io.ReadCloser, err error) error {
	if err != nil {
		return err
	}
	return body.Close()
}

// checkCalls runs every case against a fresh fake server and compares the method,
// path, query and body it saw.
func checkCalls(t *testing.T, cases []apiCall) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := newFakeServer(t, replyJSON(200, `{}`))
			if err := tc.call(context.Background(), newTestClient(t, server, "t1")); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			got := server.seen()[0]
			got.auth, got.tenant, got.agent = "", "", ""
			if got != tc.want {
				t.Errorf("request = %+v, want %+v", got, tc.want)
			}
		})
	}
}
