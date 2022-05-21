package httpsrv

import (
	"context"
	"net/http"
)

const (
	// HeaderContentLength the header name of `Content-Length`
	HeaderContentLength = "Content-Length"
	// HeaderContentType the header name of `Content-Type`
	HeaderContentType = "Content-Type"
	// MIMEApplicationJSON the application type for json
	MIMEApplicationJSON = "application/json"
	// MIMEApplicationJSONCharsetUTF8 the application type for json of utf-8 encoding
	MIMEApplicationJSONCharsetUTF8 = "application/json; charset=UTF-8"
	MIMETextPlainCharsetUTF8       = "text/plain; charset=UTF-8"
)

// ErrServerInternal indicates the server internal error
var ErrServerInternal = NewError(-1, "server internal error")

// BaseHandler is the enhanced version of ngs.BaseController
type BaseHandler struct{}

// ParseRequest parses and validates the api request

// Error writes out an error response
func (h *BaseHandler) Error(ctx context.Context, w http.ResponseWriter, err interface{}) {
	Error(ctx, w, err)
}

// OK writes out a success response without data, used typically in an `update` api.
func (h *BaseHandler) OK(ctx context.Context, w http.ResponseWriter) {
	OK(ctx, w)
}

// OKData writes out a success response with data, used typically in an `get` api.
func (h *BaseHandler) OKData(ctx context.Context, w http.ResponseWriter, data interface{}) {
	OKData(ctx, w, data)
}

// EncodeJSON is a wrapper of json.Marshal()
func (h *BaseHandler) EncodeJSON(v interface{}) ([]byte, error) {
	return EncodeJSON(v)
}

// WriteJSON writes out an object which is serialized as json.
func (h *BaseHandler) WriteJSON(w http.ResponseWriter, v interface{}) error {
	return WriteJSON(w, v)
}
