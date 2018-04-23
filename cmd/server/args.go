package main

import (
	"errors"
	"strings"

	"github.com/labstack/gommon/log"
)

type logLevel log.Lvl

func (l *logLevel) String() string {
	switch log.Lvl(*l) {
	case log.DEBUG:
		return "debug"
	case log.INFO:
		return "info"
	case log.WARN:
		return "warn"
	case log.ERROR:
		return "error"
	case log.OFF:
		return "off"
	}
	return ""
}

func (l *logLevel) Set(value string) error {
	if value == "" {
		value = "error"
	}
	switch strings.ToLower(value) {
	case "debug":
		*l = logLevel(log.DEBUG)
	case "info":
		*l = logLevel(log.INFO)
	case "warn":
		*l = logLevel(log.WARN)
	case "error":
		*l = logLevel(log.ERROR)
	case "off":
		*l = logLevel(log.OFF)
	default:
		return errors.New("invalid log level flag")
	}
	return nil
}

type snowthAddrFlag []string

func (saf *snowthAddrFlag) String() string {
	return strings.Join(*saf, ", ")
}

func (saf *snowthAddrFlag) Set(value string) error {
	*saf = append(*saf, value)
	return nil
}
