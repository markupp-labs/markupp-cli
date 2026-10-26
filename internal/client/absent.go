package client

import (
	"errors"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// CodeEndpointAbsent is the failure code of a route the server does not have.
const CodeEndpointAbsent = codeEndpointAbsent

// EndpointAbsent reports whether err is the plain text 404 of a server without the
// route, which means a server older than the command.
func EndpointAbsent(err error) bool {
	var classified *failure.Error
	return errors.As(err, &classified) && classified.Code == codeEndpointAbsent
}
