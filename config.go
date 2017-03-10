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
	minUploadDelay = 200

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

type (
	logoutput    uint8
	upload_delay uint16
	queue_size   uint16
)

type Configuration interface {
	GetMain() *MainCfg
	GetFlows() []*FlowCfg
	Validate() error
}

type MainCfg struct {
	LogLevel  string `ini:"log_level"`
	LogOutput string `ini:"log_output"`
}

type FlowCfg struct {
	Group            string       `ini:"group"`
	Stream           string       `ini:"stream"`
	SyslogFormat     string       `ini:"syslog_format"`
	CloudwatchFormat string       `ini:"cloudwatch_format"`
	Source           string       `ini:"source"`
	UploadDelay      upload_delay `ini:"upload_delay"`
	QueueSize        queue_size   `ini:"queue_size"`
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

type IniConfig struct {
	config *ini.File
}

func NewIniConfig(file string) Configuration {
	config, err := ini.Load(file)
	if err != nil {
		log.Fatalf("could not read config file %s", err)
	}
	// Remove unused default section
	config.DeleteSection(ini.DEFAULT_SECTION)
	return &IniConfig{config: config}
}

func (cfg IniConfig) GetMain() *MainCfg {
	main := new(MainCfg)
	// Set default values
	main.LogLevel = "error"
	main.LogOutput = "syslog"
	err := cfg.config.Section(mainSectionName).MapTo(main)
	if err != nil {
		log.Fatalf("could not map section %s: %s", mainSectionName, err)
	}
	return main
}

// Return all flow configurations
func (cfg IniConfig) GetFlows() (flows []*FlowCfg) {
	for _, section := range cfg.config.Sections() {
		if section.Name() != mainSectionName {
			flow := new(FlowCfg)
			// Set default values
			flow.UploadDelay = minUploadDelay
			flow.QueueSize = 50000
			err := section.MapTo(flow)
			if err != nil {
				log.Fatalf("could not map section %s: %s", mainSectionName, err)
			}
			flows = append(flows, flow)
		}
	}
	return
}

func (cfg IniConfig) Validate() error {
	if err := validateMainCfg(cfg.GetMain()); err != nil {
		return fmt.Errorf("error while validating main section: %s", err)
	}
	for _, flow := range cfg.GetFlows() {
		if err := validateFlowCfg(flow); err != nil {
			return fmt.Errorf("error while validating flow section: %s", err)
		}
	}
	return nil
}

func validateMainCfg(cfg *MainCfg) error {
	if err := validateLogLevel(cfg.LogLevel); err != nil {
		return fmt.Errorf("log_level %s", err)
	}
	if err := validateLogOutput(cfg.LogOutput); err != nil {
		return fmt.Errorf("log_output %s", err)
	}
	return nil
}

func validateFlowCfg(cfg *FlowCfg) error {
	if err := validateQueueSize(cfg.QueueSize); err != nil {
		return err
	}
	if err := validateGroup(cfg.Group); err != nil {
		return err
	}
	if err := validateStrean(cfg.Stream); err != nil {
		return err
	}
	if err := validateUploadDelay(cfg.UploadDelay); err != nil {
		return err
	}
	if err := validateSource(cfg.Source); err != nil {
		return err
	}
	if err := validateCloudwatchFormat(cfg.CloudwatchFormat); err != nil {
		return err
	}
	if err := validateSyslogFormat(cfg.SyslogFormat); err != nil {
		return err
	}
	return nil
}

// Validate source URL
func validateSource(value string) error {
	if value == "" {
		return errEmptyValue
	}
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

func validateCloudwatchFormat(value string) error {
	if value == "" {
		return errEmptyValue
	}
	_, err := template.New("").Parse(value)
	if err != nil {
		return err
	}
	return nil
}

func validateQueueSize(value queue_size) error {
	return nil
}

func validateUploadDelay(value upload_delay) error {
	if value < minUploadDelay {
		return errTooSmall
	}
	return nil
}

func validateLogOutput(value string) error {
	if value == "" {
		return errEmptyValue
	}
	if !strIn(validOutputOptions, value) {
		return errInvalidValue
	}
	return nil
}

func validateLogLevel(value string) error {
	if value == "" {
		return errEmptyValue
	}
	if !strIn(validLevelOptions, value) {
		return errInvalidValue
	}
	return nil
}

func strIn(haystack []string, needle string) bool {
	for _, elem := range haystack {
		if elem == needle {
			return true
		}
	}
	return false
}
