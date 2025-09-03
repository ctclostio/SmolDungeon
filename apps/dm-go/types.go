package main

import (
	"github.com/google/uuid"
)

// ID represents a UUID
type ID string

// NewID generates a new UUID
func NewID() ID {
	return ID(uuid.New().String())
}

// Stat represents character stats
type Stat struct {
	HP      int `json:"hp"`
	MaxHP   int `json:"maxHp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
}

// Position represents a 2D position
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Weapon represents a weapon
type Weapon struct {
	ID       ID     `json:"id"`
	Name     string `json:"name"`
	Damage   int    `json:"damage"`
	Accuracy int    `json:"accuracy"`
}

// Ability represents an ability
type Ability struct {
	ID       ID              `json:"id"`
	Name     string          `json:"name"`
	Cooldown int             `json:"cooldown"`
	Effect   string          `json:"effect"` // "damage", "heal", "buff", "debuff"
	Power    int             `json:"power"`
}

// Item represents an item
type Item struct {
	ID     ID     `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`   // "consumable", "equipment"
	Effect string `json:"effect"`
}

// Character represents a game character
type Character struct {
	ID               ID                    `json:"id"`
	Name             string                `json:"name"`
	Stats            Stat                  `json:"stats"`
	Position         Position              `json:"position"`
	Weapons          []Weapon              `json:"weapons"`
	Abilities        []Ability             `json:"abilities"`
	Items            []Item                `json:"items"`
	AbilityCooldowns map[string]int        `json:"abilityCooldowns"`
	IsPlayer         bool                  `json:"isPlayer"`
}

// Action represents a game action
type Action struct {
	Kind    string `json:"kind"`
	Attacker ID    `json:"attacker,omitempty"`
	Target   ID    `json:"target,omitempty"`
	Weapon   ID    `json:"weapon,omitempty"`
	Actor    ID    `json:"actor,omitempty"`
	Ability  ID    `json:"ability,omitempty"`
	Item     ID    `json:"item,omitempty"`
}

// Event represents a game event
type Event struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type"`
	Target ID     `json:"target,omitempty"`
	Amount int    `json:"amount,omitempty"`
	Source ID     `json:"source,omitempty"`
	Actor  ID     `json:"actor,omitempty"`
	Ability ID    `json:"ability,omitempty"`
	Item   ID     `json:"item,omitempty"`
}

// State represents the game state
type State struct {
	Round      int         `json:"round"`
	Characters []Character `json:"characters"`
	TurnOrder  []ID        `json:"turnOrder"`
	CurrentTurn int        `json:"currentTurn"`
	IsComplete bool        `json:"isComplete"`
	Winner     *string     `json:"winner,omitempty"` // "player", "enemy", "draw"
}

// Resolution represents the result of applying an action
type Resolution struct {
	Events []Event `json:"events"`
	State  State   `json:"state"`
	Logs   []string `json:"logs"`
}

// RollCheck represents a roll check request
type RollCheck struct {
	Actor ID     `json:"actor"`
	Type  string `json:"type"` // "attack", "defense", "skill", "save"
	DC    int    `json:"dc"`
}

// RollResult represents the result of a roll check
type RollResult struct {
	Roll     int  `json:"roll"`
	Modifier int  `json:"modifier"`
	Total    int  `json:"total"`
	Success  bool `json:"success"`
}