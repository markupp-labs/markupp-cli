package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"iter"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

func TestLineSanitizesAndEndsWithNewline(t *testing.T) {
	var out bytes.Buffer
	if err := NewPrinter(&out).Line("id\x1b-1\nx"); err != nil {
		t.Fatalf("Line: %v", err)
	}
	if out.String() != "id-1 x\n" {
		t.Errorf("Line wrote %q", out.String())
	}
}

func TestBlockAddsMissingNewlineOnly(t *testing.T) {
	var out bytes.Buffer
	p := NewPrinter(&out)
	if err := p.Block("a\nb"); err != nil {
		t.Fatalf("Block: %v", err)
	}
	if err := p.Block("c\n"); err != nil {
		t.Fatalf("Block: %v", err)
	}
	if out.String() != "a\nb\nc\n" {
		t.Errorf("Block wrote %q", out.String())
	}
}

func TestFieldsAlignsKeys(t *testing.T) {
	var out bytes.Buffer
	err := NewPrinter(&out).Fields(
		Field{Key: "url", Value: "http://localhost:8080"},
		Field{Key: "tenant", Value: "t\x1b1"},
	)
	if err != nil {
		t.Fatalf("Fields: %v", err)
	}
	outputtest.Golden(t, "fields", out.Bytes())
}

func TestJSONPassesBytesThroughUnchanged(t *testing.T) {
	body := `[{"id":"n1","snippet":"a \u0002b\u0003"}  ]`
	var out bytes.Buffer
	if err := NewPrinter(&out).JSON(strings.NewReader(body)); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if out.String() != body+"\n" {
		t.Errorf("JSON wrote %q, want body plus newline", out.String())
	}
}

func TestJSONKeepsExistingNewlineAndEmptyBody(t *testing.T) {
	var out bytes.Buffer
	p := NewPrinter(&out)
	if err := p.JSON(strings.NewReader("{}\n")); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if err := p.JSON(strings.NewReader("")); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if out.String() != "{}\n" {
		t.Errorf("JSON wrote %q, want %q", out.String(), "{}\n")
	}
}

func rawElements(items ...string) iter.Seq2[json.RawMessage, error] {
	return func(yield func(json.RawMessage, error) bool) {
		for _, item := range items {
			if !yield(json.RawMessage(item), nil) {
				return
			}
		}
	}
}

func TestNDJSONCompactsOneObjectPerLine(t *testing.T) {
	var out bytes.Buffer
	err := NewPrinter(&out).NDJSON(rawElements(`{ "id": "n1",
	 "path": "a.md" }`, `{"id":"n2"}`))
	if err != nil {
		t.Fatalf("NDJSON: %v", err)
	}
	outputtest.Golden(t, "ndjson", out.Bytes())
}

func TestNDJSONStopsAtTheFirstError(t *testing.T) {
	boom := errors.New("stream broke")
	elements := func(yield func(json.RawMessage, error) bool) {
		if yield(json.RawMessage(`{"id":"n1"}`), nil) {
			yield(nil, boom)
		}
	}
	var out bytes.Buffer
	if err := NewPrinter(&out).NDJSON(elements); !errors.Is(err, boom) {
		t.Fatalf("NDJSON error = %v, want %v", err, boom)
	}
	if out.String() != "{\"id\":\"n1\"}\n" {
		t.Errorf("NDJSON wrote %q", out.String())
	}
}

func TestNDJSONRejectsInvalidElement(t *testing.T) {
	var out bytes.Buffer
	err := NewPrinter(&out).NDJSON(rawElements(`{"id":`))
	if err == nil || !strings.Contains(err.Error(), `{"id":`) {
		t.Fatalf("NDJSON error = %v, want one naming the element", err)
	}
}
