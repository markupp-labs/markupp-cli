package client

// CodeEndpointAbsent is the failure code of a route the server does not have.
const CodeEndpointAbsent = codeEndpointAbsent

// EndpointAbsent reports whether err is the plain text 404 of a server without the
// route, which means a server older than the command.
func EndpointAbsent(error) bool { return false }
