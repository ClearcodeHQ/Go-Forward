package main

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-ini/ini"
)

const generalSection = "general"

type generalConfig struct {
	role string
}

var (
	errNameLooLong   = errors.New("name too long")
	errInvalidName   = errors.New("invalid name")
	errNameEmpty     = errors.New("empty name")
	errInvalidScheme = errors.New("invalid network scheme")
)

// Read and return all sections from config file
func getConfig(file string) (sections []*ini.Section, err error) {
	config, err := ini.Load(file)
	if err != nil {
		return
	}
	sections = config.Sections()
	return
}

// Read, validate and return all bons from sections
// func getBonds(sections []*ini.Section) (bonds []streamBond, err error) {
// 	for _, section := range sections {
// 		if section.Name() != generalSection {
//
// 		}
// 	}
// }

func validateSection(section *ini.Section) error {
	var required = map[string]validateKeyFunc{
		"group":  validateGroup,
		"stream": validateStrean,
		"source": validateSource,
	}
	// Check required keys
	for key, keyfunc := range required {
		if ok := section.HasKey(key); !ok {
			return fmt.Errorf("missing key %s in section %s", key, section.Name())
		}
		// Validate values
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
	if _, ok := schemes[uri.Scheme]; !ok {
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
		return errNameEmpty
	}
	if len(name) > 512 {
		return errNameLooLong
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
		return errNameEmpty
	}
	if len(name) > 512 {
		return errNameLooLong
	}
	if strings.Contains(name, ":") {
		return errInvalidName
	}
	return nil
}
