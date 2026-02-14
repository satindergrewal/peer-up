package auth

import "errors"

var (
	// ErrPeerAlreadyAuthorized is returned when trying to add a peer
	// that is already in the authorized_keys file.
	ErrPeerAlreadyAuthorized = errors.New("peer already authorized")

	// ErrPeerNotFound is returned when trying to remove a peer
	// that is not in the authorized_keys file.
	ErrPeerNotFound = errors.New("peer not found")

	// ErrInvalidPeerID is returned when a peer ID string cannot be decoded.
	ErrInvalidPeerID = errors.New("invalid peer ID")
)
