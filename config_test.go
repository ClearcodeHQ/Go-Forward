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

func Test_validateSyslogFormat_empty(t *testing.T) {
	err := validateSyslogFormat("")
	assert.Equal(t, errEmptyValue, err)
}

func Test_validateSyslogFormat_format(t *testing.T) {
	err := validateSyslogFormat("bad_format")
	assert.Equal(t, errInvalidFormat, err)
}

func Test_validateCloudwatchFormat_empty(t *testing.T) {
	err := validateCloudwatchFormat("")
	assert.Equal(t, errEmptyValue, err)
}

func Test_validateLogOutput_empty(t *testing.T) {
	err := validateLogOutput("")
	assert.Equal(t, errEmptyValue, err)
}

func Test_validateLogLevel_empty(t *testing.T) {
	err := validateLogLevel("")
	assert.Equal(t, errEmptyValue, err)
}

func Test_validateLogOutput_invalid(t *testing.T) {
	err := validateLogOutput("invalid")
	assert.Equal(t, errInvalidValue, err)
}

func Test_validateLogLevel_invalid(t *testing.T) {
	err := validateLogLevel("invalid")
	assert.Equal(t, errInvalidValue, err)
}

func Test_validateLogLevel_ok(t *testing.T) {
	for _, option := range validLevelOptions {
		assert.Nil(t, validateLogLevel(option))
	}
}

func Test_validateLogOutput_ok(t *testing.T) {
	for _, option := range validOutputOptions {
		assert.Nil(t, validateLogOutput(option))
	}
}

func Test_validateStrContains_false(t *testing.T) {
	assert.False(t, strIn([]string{}, "needle"))
}

func Test_validateStrContains_true(t *testing.T) {
	assert.True(t, strIn([]string{"needle"}, "needle"))
}

func Test_validateUploadDelay_ok(t *testing.T) {
	assert.Nil(t, validateUploadDelay(300))
}

func Test_validateUploadDelay_too_small(t *testing.T) {
	assert.Equal(t, errTooSmall, validateUploadDelay(1))
}

func Test_validateQueueSize_ok(t *testing.T) {
	assert.Nil(t, validateQueueSize(0))
}
