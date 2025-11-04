package types

// Character represents a game character
type Character struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Stats    Stat   `json:"stats"`
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"position"`
	Weapons          []Weapon          `json:"weapons"`
	Abilities        []Ability         `json:"abilities"`
	Items            []interface{}     `json:"items"`
	AbilityCooldowns map[string]int    `json:"abilityCooldowns"`
	IsPlayer         bool              `json:"isPlayer"`
}

// Stat represents character statistics
type Stat struct {
	HP      int `json:"hp"`
	MaxHP   int `json:"maxHp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
}

// Weapon represents a character's weapon
type Weapon struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Damage   int    `json:"damage"`
	Accuracy int    `json:"accuracy"`
}

// Ability represents a character's special ability
type Ability struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Cooldown int    `json:"cooldown"`
	Effect   string `json:"effect"`
	Power    int    `json:"power"`
}

// GameState represents the current game state
type GameState struct {
	Round       int         `json:"round"`
	Characters  []Character `json:"characters"`
	TurnOrder   []string    `json:"turnOrder"`
	CurrentTurn int         `json:"currentTurn"`
	IsComplete  bool        `json:"isComplete"`
	Winner      *string     `json:"winner,omitempty"`
}

// Action represents a game action
type Action struct {
	Kind     string `json:"kind"`
	Attacker string `json:"attacker,omitempty"`
	Target   string `json:"target,omitempty"`
	Weapon   string `json:"weapon,omitempty"`
	Actor    string `json:"actor,omitempty"`
	Ability  string `json:"ability,omitempty"`
}

// ActionResponse represents the server response to an action
type ActionResponse struct {
	State  GameState     `json:"state"`
	Events []interface{} `json:"events"`
	Logs   []string      `json:"logs"`
}

// CreateSessionResponse represents the response when creating a new session
type CreateSessionResponse struct {
	SessionID string    `json:"sessionId"`
	State     GameState `json:"state"`
}
