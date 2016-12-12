[![Build Status](https://travis-ci.org/ClearcodeHQ/Go-Forward.svg)](https://travis-ci.org/ClearcodeHQ/Go-Forward)
[![Coverage Status](https://coveralls.io/repos/github/ClearcodeHQ/Go-Forward/badge.svg?branch=master)](https://coveralls.io/github/ClearcodeHQ/Go-Forward?branch=master)

This program's purpose is to forward all logs received from a unix/ip socket and forward them to cloudwatch logs.

### Motivation:
* Learn go
* Small memory footprint
* No file readers
* Socket listeners

### Usage:
```
-c string
	Config file location. (default "/etc/logs_agent.cfg")
```
See [config.ini](config.ini) for possible configuration options.

### Program behaviour:
* Logs that are too old are discarded.
* Logs that exceed their allowed size are discarded.
* Incoming message timestamps are only used to set cloudwatch logs
timestamp value. They are not written in message body.
