package main

import (
	"testing"

	"github.com/go-ini/ini"
	"github.com/stretchr/testify/assert"
)

func TestValidateGroup_valid_chars(t *testing.T) {
	err := validateGroup("_-/.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	assert.Nil(t, err)
}

func TestValidateGroup_too_long(t *testing.T) {
	err := validateGroup(RandomString(518))
	assert.Equal(t, err, errNameLooLong)
}

func TestValidateGroup_empty(t *testing.T) {
	err := validateGroup("")
	assert.NotNil(t, err)
}

func TestValidateGroup_invalid_strings(t *testing.T) {
	for _, chr := range []string{",", "|", "ą"} {
		err := validateGroup(chr)
		assert.Equal(t, err, errInvalidName)
	}
}

func TestValidateStream_too_long(t *testing.T) {
	err := validateStrean(RandomString(518))
	assert.Equal(t, err, errNameLooLong)
}

func TestValidateStream_empty(t *testing.T) {
	err := validateStrean("")
	assert.NotNil(t, err)
}

func TestValidateStream_invalid_strings(t *testing.T) {
	err := validateGroup(":")
	assert.Equal(t, err, errInvalidName)
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
	i, _ := ini.Load([]byte("[asd]"))
	sec := i.Section("asd")
	err := validateSection(sec)
	assert.NotNil(t, err)
}

func TestValidateSection_ok(t *testing.T) {
	i, _ := ini.Load([]byte("[asd]"))
	sec := i.Section("asd")
	for key, val := range map[string]string{
		"group":  "some_group_name",
		"stream": "stream_name",
		"source": "udp://localhost:5514",
	} {
		sec.NewKey(key, val)
	}
	err := validateSection(sec)
	assert.Nil(t, err)
}
