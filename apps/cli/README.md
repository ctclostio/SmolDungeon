# SmolDungeon CLI

A beautiful terminal UI for SmolDungeon - a turn-based D&D combat game.

## Features

- 🎨 **Beautiful Terminal UI** - Built with Bubble Tea and lipgloss for gorgeous styling
- ⌨️ **Interactive Navigation** - Vim-style keyboard controls (↑/↓ or k/j)
- 📊 **Live Combat Stats** - Health bars, character stats, and turn indicators
- 📜 **Combat Log** - Real-time action feedback
- 🎯 **Target Selection** - Choose your enemies strategically
- 🛡️ **Multiple Actions** - Attack, Defend, or Flee
- 🎮 **Multiple Scenarios** - Play different combat encounters

## Installation

```bash
cd apps/cli
go build -o smoldungeon.exe
```

## Usage

### Start a New Game

```bash
# Play the default scenario (goblin-ambush)
./smoldungeon.exe play

# Play a specific scenario
./smoldungeon.exe play --scenario goblin-ambush
```

### List Available Scenarios

```bash
./smoldungeon.exe scenarios
```

### Custom Server URL

```bash
# Connect to a different game server
./smoldungeon.exe --server http://localhost:3000 play
```

### Help

```bash
# Show all commands
./smoldungeon.exe --help

# Show help for a specific command
./smoldungeon.exe play --help
```

## Controls

When playing:

- **↑/k** - Move cursor up
- **↓/j** - Move cursor down
- **Enter/Space** - Select action/target
- **q/Ctrl+C** - Quit game

## Architecture

The CLI is organized into clean, modular packages:

- **cmd/** - Cobra commands (play, scenarios, root)
- **ui/** - Bubble Tea models and lipgloss styles
- **api/** - HTTP client for game server communication
- **types/** - Shared data structures

## Technology Stack

- **[Cobra](https://github.com/spf13/cobra)** - Command-line interface framework
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Terminal UI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Style definitions for terminal output
- **Go 1.21+** - Programming language

## Requirements

- Go 1.21 or higher
- SmolDungeon server running (default: http://localhost:3000)

## Example Gameplay

```
⚔️  SMOLDUNGEON - Round 1  ⚔️

🗡️  YOUR PARTY
╭──────────────────────────────────────╮
│ Fighter                              │
│ HP: 38/38  ATK: 6  DEF: 4  Pos: (0,0)│
│ ██████████████████████████████       │
│ Weapon: Longsword                    │
╰──────────────────────────────────────╯

👹 ENEMIES
╭──────────────────────────────────────╮
│ Goblin Warrior ← CURRENT TURN        │
│ HP: 15/15  ATK: 4  DEF: 2  Pos: (2,0)│
│ ██████████████████████████████       │
│ Weapon: Rusty Sword                  │
╰──────────────────────────────────────╯

Fighter's Turn - Choose Action:

▶ ⚔️  Attack
  🛡️  Defend
  🏃 Flee

↑/↓: Navigate • Enter: Select • q: Quit
```

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build -o smoldungeon.exe

# Run
./smoldungeon.exe play
```

## License

Part of the SmolDungeon project.
