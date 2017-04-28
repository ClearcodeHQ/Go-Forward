package main

import (
	"math/rand"
	"time"

	"github.com/go-ini/ini"
)

type numPair struct {
	expected int
	passed   int
}

// Create a N long random string
func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func empty_ini_section() *ini.Section {
	i, _ := ini.Load([]byte(""))
	i.DeleteSection(ini.DEFAULT_SECTION)
	sec, _ := i.NewSection("fixture")
	return sec
}

var testPriorities = []TestSyslogPriority{
	{severity: logErr, facility: logMail, priority: SyslogPriority(19)},
	{severity: logEmerg, facility: logKern, priority: SyslogPriority(0)},
	{severity: logAlert, facility: logUser, priority: SyslogPriority(9)},
}

var testMessage = struct {
	raw string
	// Parsed expected fields
	severity  SyslogSeverity
	facility  SyslogFacility
	timestamp time.Time
	hostname  string
	syslogTag string
	message   string
}{
	raw:       "<86>2016-07-23T14:48:16.970210+02:00 debian sudo: pam_unix(sudo:session): session closed for user root",
	severity:  logInfo,
	facility:  logAuthpriv,
	timestamp: time.Date(2016, 7, 23, 12, 48, 16, 970210000, time.UTC),
	hostname:  "debian",
	syslogTag: "sudo:",
	message:   "pam_unix(sudo:session): session closed for user root",
}
