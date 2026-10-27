package client

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestEndpointAbsentRecognizesThePlainTextNotFound(t *testing.T) {
	server := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) { http.NotFound(w, nil) })
	_, err := newTestClient(t, server, "").ListVaults(context.Background())
	if !EndpointAbsent(err) {
		t.Errorf("EndpointAbsent(%v) = false, want true", err)
	}
	if EndpointAbsent(errors.New("other")) || EndpointAbsent(nil) {
		t.Error("EndpointAbsent is true for an unrelated error")
	}
}
