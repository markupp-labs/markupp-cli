package client

import (
	"encoding/json"
	"fmt"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// requestBody lists the schema structs sent as a request body. A new body type joins
// the union.
type requestBody interface {
	schema.NoteCreateRequest | schema.NoteUpdateRequest
}

// marshalBody encodes a schema request struct as the JSON body of a request.
func marshalBody[T requestBody](value T) ([]byte, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, failure.Wrap(failure.Internal, fmt.Sprintf("encoding request body %T", value), err)
	}
	return body, nil
}
