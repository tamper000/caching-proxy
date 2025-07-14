package errors

import "errors"

var (
	ErrOriginServer = errors.New("origin server not specified")
	ErrSecret       = errors.New("secret not specified")
)
