package frogodb

import "errors"

// ErrConnectionIsNotSet reports that no FrogoDB client is registered.
var ErrConnectionIsNotSet = errors.New("connection is not set")

// ErrConnectionConfigIsNotSet reports that FrogoDB connection configuration is missing.
var ErrConnectionConfigIsNotSet = errors.New("connection config is not set")
