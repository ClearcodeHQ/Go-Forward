package main

import (
	"fmt"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/go-ini/ini"
)

const mainSectionName = "main"

type generalConfig struct {
	role string
}

type validateKeyFunc func(value string) error

var keyValidators = map[string]validateKeyFunc{
	"group":         validateGroup,
	"stream":        validateStrean,
	"source":        validateSource,
	"syslog_format": validateSyslogFormat,
}

func getConfig(file string) (config *ini.File) {
	config, err := ini.Load(file)
	if err != nil {
		log.Fatalf("could not read config file %s", err)
	}
	// Remove unused default section
	config.DeleteSection(ini.DEFAULT_SECTION)
	for _, section := range config.Sections() {
		if section.Name() != mainSectionName {
			err := validateSection(section)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return
}

// Return all flow sections
func getFlows(config *ini.File) (flows []*ini.Section) {
	for _, section := range config.Sections() {
		if section.Name() != mainSectionName {
			flows = append(flows, section)
		}
	}
	return
}

func validateSection(section *ini.Section) error {
	for key, keyfunc := range keyValidators {
		if !section.HasKey(key) {
			return fmt.Errorf("missing key %s in section %s", key, section.Name())
		}
		if err := keyfunc(section.Key(key).String()); err != nil {
			return fmt.Errorf("bad value of %s in section %s: %s", key, section.Name(), err)
		}
	}
	return nil
}

// Validate source URL
func validateSource(value string) error {
	uri, err := url.Parse(value)
	if err != nil {
		return err
	}
	// Valid schemes
	var schemes = map[string]bool{
		"udp": true,
	}
	// Check for valid scheme
	if !schemes[uri.Scheme] {
		return errInvalidScheme
	}
	return nil
}

/*
http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogGroup.html
Log group names can be between 1 and 512 characters long.
Allowed characters are a-z, A-Z, 0-9, '_' (underscore), '-' (hyphen), '/' (forward slash), and '.' (period).
*/
func validateGroup(value string) error {
	if value == "" {
		return errEmptyValue
	}
	if len(value) > 512 {
		return errNameTooLong
	}
	for _, char := range value {
		if !strings.Contains("_-/.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", string(char)) {
			return errInvalidValue
		}
	}
	return nil
}

/*
http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogStream.html
Log stream names can be between 1 and 512 characters long
The ':' colon character is not allowed.
*/
func validateStrean(value string) error {
	if value == "" {
		return errEmptyValue
	}
	if len(value) > 512 {
		return errNameTooLong
	}
	if strings.Contains(value, ":") {
		return errInvalidValue
	}
	return nil
}

func validateSyslogFormat(value string) error {
	if value == "" {
		return errEmptyValue
	}
	if _, ok := parserFunctions[value]; !ok {
		return errInvalidFormat
	}
	return nil
}
