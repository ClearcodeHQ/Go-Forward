package main

import (
	"errors"
)

var (
	errMessageTooBig        = errors.New("message is too big.")
	errUnknownMessageFormat = errors.New("unknown syslog message format.")
	errEmptyMessage         = errors.New("message is empty.")
	errEmptyName            = errors.New("empty name")
	errNameTooLong          = errors.New("name too long")
	errInvalidName          = errors.New("invalid name")
	errInvalidScheme        = errors.New("invalid network scheme")
)
