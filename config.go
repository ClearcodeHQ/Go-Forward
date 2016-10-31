package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-ini/ini"
)

const generalSection = "main"

type generalConfig struct {
	role string
}

// Load config file and return it
func getConfig(file string) (config *ini.File, err error) {
	config, err = ini.Load(file)
	if err != nil {
		return
	}
	// Remove unused default section
	config.DeleteSection(ini.DEFAULT_SECTION)
	return
}

// Return all bonds from sections
func getBonds(config *ini.File) (bonds []streamBond) {
	for _, section := range config.Sections() {
		if section.Name() != generalSection {
			bonds = append(bonds, streamBond{
				group:  section.Key("group").String(),
				stream: section.Key("stream").String(),
				url:    section.Key("source").String(),
			})
		}
	}
	return
}

func validateSection(section *ini.Section) error {
	var required = map[string]validateKeyFunc{
		"group":  validateGroup,
		"stream": validateStrean,
		"source": validateSource,
	}
	for key, keyfunc := range required {
		if !section.HasKey(key) {
			return fmt.Errorf("missing key %s in section %s", key, section.Name())
		}
		if err := keyfunc(section.Key(key).String()); err != nil {
			return fmt.Errorf("bad value of %s in section %s: %s", key, section.Name(), err)
		}
	}
	return nil
}

type validateKeyFunc func(name string) error

// Validate source URL
func validateSource(name string) error {
	uri, err := url.Parse(name)
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
func validateGroup(name string) error {
	if name == "" {
		return errEmptyName
	}
	if len(name) > 512 {
		return errNameTooLong
	}
	for _, char := range name {
		if !strings.Contains("_-/.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", string(char)) {
			return errInvalidName
		}
	}
	return nil
}

/*
http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogStream.html
Log stream names can be between 1 and 512 characters long
The ':' colon character is not allowed.
*/
func validateStrean(name string) error {
	if name == "" {
		return errEmptyName
	}
	if len(name) > 512 {
		return errNameTooLong
	}
	if strings.Contains(name, ":") {
		return errInvalidName
	}
	return nil
}
