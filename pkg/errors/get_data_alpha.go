package errors

import "errors"

var (
	ErrSymbolRequired       = errors.New("symbol is required")
	ErrAPIKeyRequired       = errors.New("api key is required")
	ErrBaseURLRequired      = errors.New("base url is required")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)
