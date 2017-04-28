package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Parse_RFC3164_Severity(t *testing.T) {
	msg := "<86>Jul 23 14:48:16 debian sudo: pam_unix(sudo:session): session closed for user root"
	parsed, err := parseRFC3164(msg)
	assert.Equal(t, logInfo, parsed.Severity)
	assert.Equal(t, logAuthpriv, parsed.Facility)
	assert.Equal(t, time.Month(7), parsed.timestamp.Month())
	assert.Equal(t, "debian", parsed.Hostname)
	assert.Equal(t, "sudo:", parsed.Syslogtag)
	assert.Equal(t, "pam_unix(sudo:session): session closed for user root", parsed.Message)
	assert.Nil(t, err)
}

func Test_parseRFC3164_empty(t *testing.T) {
	msg := "<86>Jul 23 14:48:16 debian sudo:"
	_, err := parseRFC3164(msg)
	assert.NotNil(t, err)
}

func TestUnknownMessage(t *testing.T) {
	bad_messages := []string{
		"kfjlsdkfdlsjdlfgkdlsfghsdlfgkh",
		"<888>dsfdsfdsgsgd",
		"<aa>bla bla#@$@#4",
		"<84>bla bla#@$@#4",
	}
	for _, msg := range bad_messages {
		_, err := parseRFC3164(msg)
		assert.Equal(t, errUnknownMessageFormat, err)
	}
}

func Benchmark_parseRCF3164(b *testing.B) {
	msg := "<86>Jul 23 14:48:16 debian sudo: pam_unix(sudo:session): session closed for user root"
	for n := 0; n < b.N; n++ {
		parseRFC3164(msg)
	}
}
