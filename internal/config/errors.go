package config

import "errors"

var (
	// ErrConfigNotFound is returned when no config file is found
	// at the specified path or in any of the search paths.
	ErrConfigNotFound = errors.New("config file not found")

	// ErrConfigVersionTooNew is returned when a config file has a version
	// newer than what this binary supports.
	ErrConfigVersionTooNew = errors.New("config version too new")
)
