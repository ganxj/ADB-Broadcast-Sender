package app

import "errors"

var (
	ErrNoDeviceSelected    = errors.New("no device selected")
	ErrDeviceNotConnected  = errors.New("device not connected")
	ErrBroadcastFailed     = errors.New("broadcast failed")
	ErrConfigLoadFailed    = errors.New("failed to load configuration")
)
