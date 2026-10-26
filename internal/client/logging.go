package client

import (
	"log/slog"
	"net/http"
	"time"
)

// loggingTransport logs one debug record per request with the method, the path, the
// status and the duration. It never logs headers, the query or bodies, which carry
// the token, search terms and note content.
type loggingTransport struct {
	next   http.RoundTripper
	logger *slog.Logger
}

func (l *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	started := time.Now()
	resp, err := l.next.RoundTrip(req)
	attrs := []slog.Attr{
		slog.String("method", req.Method),
		slog.String("host", req.URL.Host),
		slog.String("path", req.URL.Path),
		slog.Int64("duration_ms", time.Since(started).Milliseconds()),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	} else {
		attrs = append(attrs, slog.Int("status", resp.StatusCode))
	}
	l.logger.LogAttrs(req.Context(), slog.LevelDebug, "http request", attrs...)
	return resp, err
}
