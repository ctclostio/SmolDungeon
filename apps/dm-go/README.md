# SmolDungeon DM Server (Go)

A high-performance Dungeon Master server for SmolDungeon, built with Go and Fiber.

## Features

- **High Performance**: Built with Go for excellent concurrency and low latency
- **REST API**: Compatible with the existing SmolDungeon API
- **SQLite Persistence**: Event sourcing with snapshots for game state
- **LLM Integration**: OpenAI-compatible API for narration and AI enemy actions
- **Real-time Ready**: Built on Fiber for potential WebSocket support

## Quick Start

### Prerequisites

- Go 1.21 or later
- SQLite3 development libraries

### Installation

1. Install dependencies:
```bash
go mod download
```

2. Build the server:
```bash
go build -o dm-server .
```

3. Run the server:
```bash
./dm-server
```

Or use the npm scripts:
```bash
pnpm run dev:dm      # Development mode
pnpm run build:dm    # Build for production
```

## Configuration

Configure the server using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Server port |
| `DB_PATH` | `./dm-server.db` | SQLite database path |
| `LLM_API_KEY` | `dummy-key` | OpenAI API key |
| `LLM_BASE_URL` | `` | Custom LLM base URL |
| `LLM_MODEL` | `gpt-3.5-turbo` | LLM model to use |
| `LLM_MAX_TOKENS` | `150` | Max tokens for LLM responses |
| `LLM_TEMPERATURE` | `0.7` | LLM temperature setting |

## API Endpoints

### Tools

- `POST /tools/get_state_summary` - Get a text summary of the game state
- `POST /tools/roll_check` - Perform a dice roll check
- `POST /tools/apply_action` - Apply a game action and get the resolution

### Sessions

- `GET /health` - Health check
- `GET /sessions` - List active sessions
- `POST /sessions` - Create a new session
- `GET /sessions/:sessionId` - Get session state

### Headers

- `session-id` - Optional header for associating requests with game sessions

## Architecture

The server implements an event-sourced architecture:

- **Events**: Immutable records of what happened in the game
- **Snapshots**: Periodic saves of the complete game state
- **State**: Reconstructed from events and snapshots for each request

## Migration from Node.js

This Go server is a drop-in replacement for the original Node.js/Fastify DM server. All API endpoints are compatible, so existing clients will work without changes.

### Performance Benefits

- **Concurrency**: Go's goroutines handle thousands of concurrent sessions
- **Memory**: Lower memory footprint than Node.js
- **Deployment**: Single binary deployment, no runtime dependencies
- **Startup**: Faster cold starts

## Development

### Project Structure

```
apps/dm-go/
├── main.go          # Server entry point and API routes
├── types.go         # Data structures and types
├── core.go          # Game logic (ported from TypeScript)
├── database.go      # SQLite persistence layer
├── rng.go           # Random number generation
├── llm.go           # LLM client for AI features
├── go.mod           # Go module definition
└── README.md        # This file
```

### Testing

Run the test suite:
```bash
go test ./...
```

### Building for Production

```bash
go build -ldflags="-s -w" -o dm-server .
```

This creates a stripped binary optimized for production.