package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

func TestTableWritesFixedWidthRows(t *testing.T) {
	var out bytes.Buffer
	table := NewPrinter(&out).Table(
		Column{Name: "ID", Width: 4},
		Column{Name: "PATH", Width: 10},
		Column{Name: "TITLE"},
	)
	rows := [][]string{
		{"n1", "a.md", "Alpha"},
		{"n2", "very/long/path.md", "Beta\nwith newline"},
		{"n3", "", "empty path"},
	}
	if err := table.Header(); err != nil {
		t.Fatalf("Header: %v", err)
	}
	for _, row := range rows {
		if err := table.Row(row...); err != nil {
			t.Fatalf("Row(%q): %v", row, err)
		}
	}
	outputtest.Golden(t, "table", out.Bytes())
}

func TestTableRowIsWrittenBeforeTheNextOne(t *testing.T) {
	var out bytes.Buffer
	table := NewPrinter(&out).Table(Column{Name: "ID", Width: 2}, Column{Name: "X"})
	if err := table.Row("n1", "x"); err != nil {
		t.Fatalf("Row: %v", err)
	}
	if out.String() != "n1  x\n" {
		t.Errorf("after one Row the output is %q, want %q", out.String(), "n1  x\n")
	}
}

func TestTableRowRejectsWrongCellCount(t *testing.T) {
	table := NewPrinter(&bytes.Buffer{}).Table(Column{Name: "ID", Width: 2}, Column{Name: "X"})
	err := table.Row("only one")
	if err == nil || !strings.Contains(err.Error(), "1 cells") {
		t.Fatalf("Row error = %v, want one naming the cell count", err)
	}
}
