# SmolDungeon Quick Start Guide

SmolDungeon currently ships with two runnable clients:

- Go web server with server-rendered HTML templates
- Standalone terminal CLI

## Prerequisites

- Go 1.21+
- Optional: Node.js/npm if you want to use the root `package.json` shortcuts

## Web UI

Run the Go server:

```bash
cd apps/dm-go
go run .
```

Open http://localhost:3000.

From the repository root, the same server can be started with:

```bash
npm run dev
```

## CLI

Run the standalone terminal client:

```bash
cd apps/cli
go run . play
```

From the repository root:

```bash
npm run play
```

Useful CLI commands:

```bash
go run . scenarios
go run . play --scenario goblin-ambush
go run . saves
go run . continue
```

## Tests

```bash
cd apps/dm-go
go test ./...

cd ../cli
go test ./...
```

Or from the repository root for the backend:

```bash
npm test
```

## Configuration

The Go server reads these environment variables:

```bash
PORT=3000
DB_PATH=./dm-server.db
SCENARIO_DIR=../../scenarios
LLM_API_KEY=your-openai-key
LLM_MODEL=gpt-3.5-turbo
```

LLM features are optional. Combat still runs without a real API key.
