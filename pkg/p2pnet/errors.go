package p2pnet

import "errors"

var (
	// ErrServiceAlreadyRegistered is returned when trying to register a service
	// that already exists in the registry.
	ErrServiceAlreadyRegistered = errors.New("service already registered")

	// ErrNameNotFound is returned when a name cannot be resolved to a peer ID
	// and is not a valid peer ID itself.
	ErrNameNotFound = errors.New("name not found")
)
