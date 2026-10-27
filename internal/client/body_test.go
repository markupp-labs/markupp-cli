package client

import (
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

func TestMarshalBodyEncodesTheSchemaStruct(t *testing.T) {
	got, err := marshalBody(schema.NoteCreateRequest{Path: "a.md", Content: "<b>"})
	if err != nil || string(got) != "{\"path\":\"a.md\",\"content\":\"\\u003cb\\u003e\"}" {
		t.Errorf("marshalBody = %s, %v", got, err)
	}
}
