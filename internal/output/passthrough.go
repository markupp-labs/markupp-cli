package output

import (
	"fmt"
	"io"
)

// lastByteWriter remembers the last byte it forwarded, so the passthrough knows
// whether the body ended with a newline without holding the body.
type lastByteWriter struct {
	out     io.Writer
	last    byte
	written int64
}

func (w *lastByteWriter) Write(chunk []byte) (int, error) {
	n, err := w.out.Write(chunk)
	if n > 0 {
		w.last = chunk[n-1]
		w.written += int64(n)
	}
	return n, err
}

// JSON copies a JSON body unchanged as it streams in, adding a final newline when the
// body lacks one. An empty body writes nothing.
func (p *Printer) JSON(body io.Reader) error {
	tracker := &lastByteWriter{out: p.out}
	if _, err := io.Copy(tracker, body); err != nil {
		return fmt.Errorf("copying JSON body to stdout: %w", err)
	}
	if tracker.written == 0 || tracker.last == '\n' {
		return nil
	}
	return p.write("\n")
}
