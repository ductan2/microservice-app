package types

import "net/http"

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}
