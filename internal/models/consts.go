package models

import "sync"

var (
	Accounts   map[string]Account
	AccountsMu sync.RWMutex

	Grobbles   map[string][]Grobble
	GrobblesMu sync.RWMutex
)

const (
	TokensFile   = "tokens.json"
	GrobblesFile = "grobbles.json"
)