package models

import "sync"

var (
	Accounts   map[string]Account
	AccountsMu sync.RWMutex
)

const (
	TokensFile   = "tokens.json"
	GrobblesFile = "grobbles.db"
)