# How to Play SmolDungeon

SmolDungeon can be played through the Go web UI or the standalone CLI.

## Web UI

```bash
cd apps/dm-go
go run .
```

Open http://localhost:3000, choose a scenario, and start a game.

## CLI

```bash
cd apps/cli
go run . play
```

Or from the repository root:

```bash
npm run play
```

Useful commands:

```bash
go run . scenarios
go run . play --scenario goblin-ambush
go run . saves
go run . continue
```

## Controls

- Up/down or `k`/`j`: navigate
- Enter or space: select
- `ctrl+s`: save
- `q`: quit

## Gameplay Tips

- Attack removes enemies fastest, but defending can buy time when health is low.
- Flee can end combat, but it can fail.
- Turn order is based on Speed plus a d20 initiative roll.
- The CLI and web UI automatically resolve enemy turns, so input is only needed on player turns.

## Development Commands

From the project root:

```bash
npm run dev        # Start Go web server
npm run play       # Start CLI game
npm run build      # Build Go web server
npm run build:cli  # Build CLI binary
npm test           # Run backend tests
```
