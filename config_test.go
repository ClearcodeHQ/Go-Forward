package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGroup_valid_chars(t *testing.T) {
	err := validateGroup("_-/.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	assert.Nil(t, err)
}

func TestValidateGroup_too_long(t *testing.T) {
	err := validateGroup(RandomString(518))
	assert.Equal(t, err, errNameTooLong)
}

func TestValidateGroup_empty(t *testing.T) {
	err := validateGroup("")
	assert.Equal(t, errEmptyValue, err)
}

func TestValidateGroup_invalid_strings(t *testing.T) {
	for _, chr := range []string{",", "|", "ą"} {
		err := validateGroup(chr)
		assert.Equal(t, err, errInvalidValue)
	}
}

func TestValidateStream_too_long(t *testing.T) {
	err := validateStrean(RandomString(518))
	assert.Equal(t, err, errNameTooLong)
}

func TestValidateStream_empty(t *testing.T) {
	err := validateStrean("")
	assert.Equal(t, errEmptyValue, err)
}

func TestValidateStream_invalid_strings(t *testing.T) {
	err := validateGroup(":")
	assert.Equal(t, err, errInvalidValue)
}

func TestValidateSource_ok(t *testing.T) {
	for _, uri := range []string{
		"udp://localhost:5514",
	} {
		err := validateSource(uri)
		assert.Nil(t, err)
	}
}

func TestValidateSource_error(t *testing.T) {
	for uri, expected := range map[string]error{
		"tcp://localhost:5514": errInvalidScheme,
	} {
		err := validateSource(uri)
		assert.Equal(t, err, expected)
	}
}

func TestValidateSection_missing_key(t *testing.T) {
	sec := fixture_valid_config().Section("valid")
	sec.DeleteKey("group")
	err := validateSection(sec)
	assert.NotNil(t, err)
}

func TestValidateSection_ok(t *testing.T) {
	sec := fixture_valid_config().Section("valid")
	err := validateSection(sec)
	assert.Nil(t, err)
}

func Test_validateSyslogFormat_empty(t *testing.T) {
	err := validateSyslogFormat("")
	assert.Equal(t, errEmptyValue, err)
}

func Test_validateSyslogFormat_format(t *testing.T) {
	err := validateSyslogFormat("bad_format")
	assert.Equal(t, errInvalidFormat, err)
}
