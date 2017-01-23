package main

import (
	"fmt"
	"net/url"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/go-ini/ini"
)

const (
	mainSectionName = "main"

	logOutputKey = "log_output"
	logLevelKey  = "log_level"

	sourceKey           = "source"
	groupKey            = "group"
	streamKey           = "stream"
	cloudwatchFormatKey = "cloudwatch_format"
	syslogFormatKey     = "syslog_format"
	queueSizeKey        = "queue_size"
	uploadDelayKey      = "upload_delay"

	debugLevelOption = "debug"
	infoLevelOption  = "info"
	errorLevelOption = "error"

	syslogOutputOption = "syslog"
	nullOutputOption   = "null"
	stdoutOutputOption = "stdout"
	stderrOutputOption = "stderr"
)

type logoutput uint8

type mainConfig struct {
	logLevel  log.Level
	logOutput logoutput
}

type flowCfg struct {
	dst         *destination
	syslogFn    syslogParser
	format      *template.Template
	recv        receiver
	uploadDelay uint16
	queueSize   uint16
}

const (
	stdErr logoutput = iota
	stdOut
	sysLog
	null
)

var strToOutput = map[string]logoutput{
	syslogOutputOption: sysLog,
	nullOutputOption:   null,
	stdoutOutputOption: stdOut,
	stderrOutputOption: stdErr,
}

var validOutputOptions = []string{
	syslogOutputOption,
	nullOutputOption,
	stdoutOutputOption,
	stderrOutputOption,
}

var strToLevel = map[string]log.Level{
	debugLevelOption: log.DebugLevel,
	infoLevelOption:  log.InfoLevel,
	errorLevelOption: log.ErrorLevel,
}

var validLevelOptions = []string{
	debugLevelOption,
	infoLevelOption,
	errorLevelOption,
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
			err := validateFlowSection(section)
			if err != nil {
				log.Fatal(err)
			}
		}
		if section.Name() == mainSectionName {
			err := validateMainSection(section)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return
}

// Return main config from section
func getMainConfig(config *ini.File) mainConfig {
	section := config.Section(mainSectionName)
	return mainConfig{
		logLevel:  strToLevel[section.Key(logLevelKey).In(errorLevelOption, validLevelOptions)],
		logOutput: strToOutput[section.Key(logOutputKey).In(syslogOutputOption, validOutputOptions)],
	}
}

// Return all flow configurations
func getFlows(config *ini.File) (flows []flowCfg) {
	for _, section := range config.Sections() {
		if section.Name() != mainSectionName {
			format, _ := template.New("").Parse(section.Key(cloudwatchFormatKey).String())
			cfg := flowCfg{
				dst: &destination{
					group:  section.Key(groupKey).String(),
					stream: section.Key(streamKey).String(),
				},
				recv:     newReceiver(section.Key(sourceKey).String()),
				format:   format,
				syslogFn: parserFunctions[section.Key(syslogFormatKey).String()],
			}
			flows = append(flows, cfg)
		}
	}
	return
}

func validateMainSection(section *ini.Section) error {
	for key, keyfunc := range mainKeyValidators {
		if !section.HasKey(key) {
			return fmt.Errorf("missing key %s in section %s", key, section.Name())
		}
		if err := keyfunc(section.Key(key).String()); err != nil {
			return fmt.Errorf("bad value of %s in section %s: %s", key, section.Name(), err)
		}
	}
	return nil
}

func validateFlowSection(section *ini.Section) error {
}

// Validate source URL
func validateSource(section *ini.Section) error {
	if !section.HasKey(sourceKey) {
		return fmt.Errorf("missing key %s in section %s", sourceKey, section.Name())
	}
	value := section.Key(sourceKey).String()
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
func validateGroup(section *ini.Section) error {
	if !section.HasKey(groupKey) {
		return fmt.Errorf("missing key %s in section %s", groupKey, section.Name())
	}
	value := section.Key(groupKey).String()
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
func validateStrean(section *ini.Section) error {
	if !section.HasKey(streamKey) {
		return fmt.Errorf("missing key %s in section %s", streamKey, section.Name())
	}
	value := section.Key(streamKey).String()
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

func validateSyslogFormat(section *ini.Section) error {
	if !section.HasKey(syslogFormatKey) {
		return fmt.Errorf("missing key %s in section %s", syslogFormatKey, section.Name())
	}
	value := section.Key(syslogFormatKey).String()
	if value == "" {
		return errEmptyValue
	}
	if _, ok := parserFunctions[value]; !ok {
		return errInvalidFormat
	}
	return nil
}

func validateCloudwatchFormat(section *ini.Section) error {
	if !section.HasKey(cloudwatchFormatKey) {
		return fmt.Errorf("missing key %s in section %s", cloudwatchFormatKey, section.Name())
	}
	value := section.Key(cloudwatchFormatKey).String()
	if value == "" {
		return errEmptyValue
	}
	_, err := template.New("").Parse(value)
	if err != nil {
		return err
	}
	return nil
}

func validateQueueSize(section *ini.Section) error {
	if section.HasKey(sourceKey) {
		value, _ := section.Key(queueSizeKey).Uint()
		if value < 0 {
			return errInvalidValue
		}
	}
	return nil
}

func validateUploadDelay(value uint) error {
	if value < 200 {
		return errTooSmall
	}
	return nil
}

func validateLogOutput(value string) error {
	if value == "" {
		return errEmptyValue
	}
	if !strContains(validOutputOptions, value) {
		return errInvalidValue
	}
	return nil
}

func validateLogLevel(value string) error {
	if value == "" {
		return errEmptyValue
	}
	if !strContains(validLevelOptions, value) {
		return errInvalidValue
	}
	return nil
}

func strContains(haystack []string, needle string) bool {
	for _, elem := range haystack {
		if elem == needle {
			return true
		}
	}
	return false
}
