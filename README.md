# Smol Dungeon

A TypeScript-first combat micro-RPG with deterministic core, Ink TUI, and LLM "Director" that mutates state only via tools. Fast to ship, easy for agents to edit, model-agnostic.

## Features

- 🎲 **Deterministic Combat Engine**: Pure functions with seeded RNG
- 🖥️ **Ink TUI Interface**: Rich command-line game experience
- 🤖 **LLM Director**: AI-powered narration and enemy behavior
- 📊 **Event Sourcing**: SQLite + Drizzle for persistent game state
- 🧪 **Property Testing**: Comprehensive test suite with fast-check
- 📦 **Monorepo Structure**: Clean separation of concerns
- 🔧 **Model Agnostic**: Works with OpenAI, llama.cpp, vLLM

## Quick Start

```bash
# Install dependencies
pnpm install

# Build all packages
pnpm build

# Run the CLI game
pnpm --filter @smol-dungeon/cli dev

# Or start the DM server
pnpm --filter @smol-dungeon/dm dev

# Run tests
pnpm test
```

## Project Structure

```
/apps
  /cli        # Ink TUI game interface
  /dm         # HTTP server with DM tools
/packages
  /core       # Pure combat engine
  /schema     # Zod type definitions
  /persistence # SQLite + Drizzle event store
  /adapters   # LLM client & scenario loader
/scenarios    # YAML scenario definitions
/tests        # Property & snapshot tests
```

## Core Contracts

### Action Types
```typescript
type Action =
  | { kind: "Attack"; attacker: Id; target: Id; weapon: Id }
  | { kind: "Defend"; actor: Id }
  | { kind: "Ability"; actor: Id; ability: Id; target?: Id }
  | { kind: "UseItem"; actor: Id; item: Id }
  | { kind: "Flee"; actor: Id };
```

### Resolution
```typescript
type Resolution = { 
  events: Event[]; 
  state: State; 
  logs: string[] 
};
```

### Core Rule
```typescript
// Pure/Deterministic: (state, action, seed) -> Resolution
function applyAction(state: State, action: Action, seed: number): Resolution
```

## DM Tools API

The DM server exposes three HTTP endpoints for LLM integration:

### GET `/tools/get_state_summary`
```typescript
POST /tools/get_state_summary
Body: { state: State }
Response: { summary: string }
```

### POST `/tools/roll_check`
```typescript
POST /tools/roll_check  
Body: { actor: Id; type: "attack"|"defense"|"skill"|"save"; dc: number }
Response: { roll: number; modifier: number; total: number; success: boolean }
```

### POST `/tools/apply_action`
```typescript
POST /tools/apply_action
Body: { state: State; action: Action; seed: number }
Response: { events: Event[]; state: State; logs: string[] }
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# LLM Configuration
MODEL=llama3.1:8b                    # Model name
CTX=4096                             # Context window
OPENAI_BASE_URL=http://localhost:11434/v1  # API endpoint

# Database
DB_PATH=./smol-dungeon.db            # SQLite database path

# Server
PORT=3000                            # DM server port

# Game
SEED=12345                           # Random seed for determinism
```

## Gameplay

### Turn-Based Combat
- **Attack**: Choose weapon and target
- **Defend**: Increase defense for the round
- **Ability**: Use special abilities (with cooldowns)
- **Use Item**: Consume potions, equipment
- **Flee**: Attempt to escape combat

### Character Stats
- **HP**: Health points (0 = defeated)
- **Attack**: Influences hit chance and damage
- **Defense**: Reduces incoming damage
- **Speed**: Determines initiative order

### Pre-made Classes
- **Fighter**: High HP, powerful attacks, defensive abilities
- **Cleric**: Healing magic, undead turning, holy weapons
- **Rogue**: High speed, sneak attacks, evasion skills

## Scenarios

Scenarios are defined in YAML format:

```yaml
name: "Goblin Ambush"
description: "Goblins attack on a forest path"
context: "Goblins leap from the bushes!"

players:
  - name: "Fighter"
    stats: { hp: 30, maxHp: 30, attack: 6, defense: 4, speed: 3 }
    weapons: [...]
    abilities: [...]

enemies:
  - name: "Goblin Warrior"
    stats: { hp: 15, maxHp: 15, attack: 4, defense: 2, speed: 5 }
    # ...
```

## Testing

### Property Tests
```bash
# Run property-based tests with fast-check
pnpm test tests/combat.test.ts
```

Tests verify:
- Non-negative damage values
- Combat termination within reasonable rounds  
- Deterministic behavior with same seed
- Valid turn order maintenance

### Snapshot Tests
```bash
# Run snapshot tests for consistent narration
pnpm test tests/snapshots.test.ts
```

Captures:
- Combat logs for fixed seed/scene
- Event sequences
- State transitions

## Development

### Build System
```bash
pnpm build           # Build all packages
pnpm dev             # Watch mode for all packages
pnpm typecheck       # Type checking
pnpm lint            # ESLint
```

### Package Scripts
```bash
# CLI app
pnpm --filter @smol-dungeon/cli dev

# DM server  
pnpm --filter @smol-dungeon/dm dev

# Individual packages
pnpm --filter @smol-dungeon/core build
```

## Architecture

### Event Sourcing
- All game events stored as append-only log
- State snapshots saved every N rounds
- Full audit trail and replay capability

### Deterministic Core
- Seeded RNG for reproducible outcomes
- Pure functions with no side effects
- Testable combat resolution

### LLM Integration
- OpenAI-compatible API client
- Model-agnostic design (works with llama.cpp, vLLM)
- Tools-based state mutation only

## License

MIT