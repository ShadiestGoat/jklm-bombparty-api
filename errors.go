package main

import "errors"

var (
	ErrNotConnected = errors.New("NotConnected")
	ErrNoRoom       = errors.New("NoSuchRoom")
)
