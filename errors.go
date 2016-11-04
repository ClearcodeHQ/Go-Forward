package main

import (
	"errors"
)

var (
	errMessageTooBig        = errors.New("message is too big")
	errUnknownMessageFormat = errors.New("unknown syslog message format")
	errEmptyMessage         = errors.New("message is empty")
	errEmptyValue           = errors.New("empty value")
	errInvalidValue         = errors.New("invalid value")
	errNameTooLong          = errors.New("name too long")
	errInvalidScheme        = errors.New("invalid network scheme")
	errInvalidFormat        = errors.New("invalid format")
)
