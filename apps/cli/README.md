# SmolDungeon CLI

A beautiful standalone terminal UI for SmolDungeon - turn-based D&D combat with **no server required!**

## Features

- 🎨 **Beautiful Terminal UI** - Built with Bubble Tea and lipgloss for gorgeous styling
- ⌨️ **Interactive Navigation** - Vim-style keyboard controls (↑/↓ or k/j)
- 📊 **Live Combat Stats** - Health bars, character stats, and turn indicators
- 📜 **Combat Log** - Real-time action feedback
- 🎯 **Target Selection** - Choose your enemies strategically
- 🛡️ **Multiple Actions** - Attack, Defend, or Flee
- 🎮 **Multiple Scenarios** - Play different combat encounters
- 💾 **Auto-Save** - Automatically saves progress after each turn
- 📁 **Save Management** - List, load, and continue saved games
- 🔌 **100% Offline** - No server, no internet connection needed!

## Installation

```bash
cd apps/cli
go build -o smoldungeon.exe
```

## Quick Start

```bash
# Start a new game (default scenario)
./smoldungeon.exe play

# Play a specific scenario
./smoldungeon.exe play --scenario goblin-ambush

# Continue your last game
./smoldungeon.exe continue

# List all saved games
./smoldungeon.exe saves

# List available scenarios
./smoldungeon.exe scenarios
```

## Commands

### `play` - Start New Game

```bash
# Start with default scenario (goblin-ambush)
./smoldungeon.exe play

# Choose a specific scenario
./smoldungeon.exe play --scenario goblin-ambush

# Disable auto-save
./smoldungeon.exe play --auto-save=false
```

### `continue` - Resume Saved Game

```bash
# Continue most recent save
./smoldungeon.exe continue

# Continue a specific save
./smoldungeon.exe continue --save auto_goblin-ambush
```

### `saves` - List Saved Games

```bash
# Show all save files with details
./smoldungeon.exe saves
```

### `scenarios` - List Scenarios

```bash
# Show available scenarios
./smoldungeon.exe scenarios
```

## Controls

When playing:

- **↑/k** - Move cursor up
- **↓/j** - Move cursor down
- **Enter/Space** - Select action/target
- **Ctrl+S** - Manual save
- **q/Ctrl+C** - Quit game (auto-saves if enabled)

## Global Flags

- `--save-dir <path>` - Custom save directory (default: `~/.smoldungeon/saves`)
- `--scenario-dir <path>` - Custom scenario directory (default: `./scenarios`)
- `--auto-save` - Enable/disable auto-save (default: `true`)

## Architecture

The CLI is completely self-contained with no external dependencies except scenario files:

```
apps/cli/
├── cmd/                    # Cobra commands
│   ├── root.go            # Base command & global flags
│   ├── play.go            # New game command
│   ├── continue.go        # Load save command
│   ├── saves.go           # List saves command
│   └── scenarios.go       # List scenarios command
├── ui/                     # Bubble Tea UI
│   ├── game.go            # Interactive game model
│   └── styles.go          # lipgloss styling
├── engine/                 # Game engine
│   ├── core.go            # Core game logic
│   ├── ai.go              # AI decision making
│   ├── types.go           # Data structures
│   ├── rng.go             # Random number generation
│   └── scenario.go        # Scenario loading
├── saves/                  # Save system
│   └── saves.go           # Save/load functionality
├── scenarios/              # Scenario files
│   └── goblin-ambush.yaml # Example scenario
├── main.go                # Entry point
└── go.mod                 # Dependencies
```

## Technology Stack

✅ **[Cobra](https://github.com/spf13/cobra)** - CLI framework
✅ **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Terminal UI framework
✅ **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Terminal styling
✅ **[yaml.v3](https://gopkg.in/yaml.v3)** - YAML parsing
✅ **Go 1.21+** - Programming language

## Save Files

Save files are stored as JSON in the save directory:

- **Location**: `~/.smoldungeon/saves/` (configurable)
- **Format**: JSON
- **Auto-save**: One per scenario (`auto_<scenario-name>.json`)
- **Manual**: Named saves (coming soon)

Example save structure:
```json
{
  "name": "auto_goblin-ambush",
  "scenarioName": "goblin-ambush",
  "state": { ... },
  "logs": [ ... ],
  "savedAt": "2025-01-03T21:30:00Z"
}
```

## Scenario Files

Scenarios are YAML files defining characters, enemies, and stats:

```yaml
name: "Goblin Ambush"
description: "A group of goblins attacks the party on a forest path"

players:
  - name: "Fighter"
    stats:
      hp: 38
      maxHp: 38
      attack: 6
      defense: 4
      speed: 5
    position:
      x: 0
      y: 0
    weapons:
      - name: "Longsword"
        damage: 8
        accuracy: 85

enemies:
  - name: "Goblin Warrior"
    stats:
      hp: 15
      maxHp: 15
      attack: 4
      defense: 2
      speed: 4
    position:
      x: 2
      y: 0
    weapons:
      - name: "Rusty Sword"
        damage: 4
        accuracy: 75
```

## Example Gameplay

```
⚔️  SMOLDUNGEON - Round 1  ⚔️

🗡️  YOUR PARTY
╭──────────────────────────────────────╮
│ Fighter ← CURRENT TURN               │
│ HP: 38/38  ATK: 6  DEF: 4  Pos: (0,0)│
│ ██████████████████████████████       │
│ Weapon: Longsword                    │
╰──────────────────────────────────────╯

👹 ENEMIES
╭──────────────────────────────────────╮
│ Goblin Warrior                       │
│ HP: 15/15  ATK: 4  DEF: 2  Pos: (2,0)│
│ ██████████████████████████████       │
│ Weapon: Rusty Sword                  │
╰──────────────────────────────────────╯

Fighter's Turn - Choose Action:

▶ ⚔️  Attack
  🛡️  Defend
  🏃 Flee

↑/↓: Navigate • Enter: Select • Ctrl+S: Save • q: Quit
```

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build -o smoldungeon.exe

# Run
./smoldungeon.exe play

# Run with custom settings
./smoldungeon.exe play --scenario-dir ./my-scenarios --save-dir ./my-saves
```

## What's Different from Server Version?

This is a **standalone, simplified version** with:

- ✅ No server required
- ✅ No HTTP/WebSocket communication
- ✅ Local JSON save files instead of SQLite database
- ✅ Embedded game engine
- ✅ Faster startup and response
- ✅ Works completely offline
- ✅ Simpler architecture

## License

Part of the SmolDungeon project.
