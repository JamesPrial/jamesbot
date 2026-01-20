// Package control provides the HTTP control API for JamesBot.
package control

import "errors"

// ErrRuleNotFound is returned when a rule is not found.
var ErrRuleNotFound = errors.New("rule not found")

// Stats contains bot statistics.
type Stats struct {
	Uptime           string `json:"uptime"`
	StartTime        int64  `json:"start_time"`
	CommandsExecuted int64  `json:"commands_executed"`
	GuildCount       int    `json:"guild_count"`
	ActiveRules      int    `json:"active_rules"`
}

// Rule represents a moderation rule.
type Rule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Key         string `json:"key"`
	Value       string `json:"value"`
}

// BotInfo is the interface that the bot must implement to provide info to the control API.
type BotInfo interface {
	Stats() *Stats
	Rules() []Rule
	SetRule(name, key, value string) error
}
