# 🚀 SmolDungeon Go

**High-performance turn-based combat game powered by Go**

SmolDungeon Go is a clean, fast implementation of SmolDungeon featuring:
- ⚡ **Go Fiber Backend** - High-performance REST API with HTML templates
- 🗄️ **SQLite Database** - Event-sourced persistence
- 🤖 **LLM Integration** - AI-powered enemy actions
- 🏗️ **Clean Architecture** - Single-binary deployment

## Quick Start

### Prerequisites
- Go 1.21+

### Run the Server
```bash
# Development mode
go run ./apps/dm-go

# Or use npm script
npm run dev
```

This will start the server at: http://localhost:3000

## Project Structure

```
smol-dungeon-go/
├── apps/
│   ├── dm-go/          # Go Fiber backend with HTML templates
│   │   ├── main.go     # Server entry point
│   │   ├── types.go    # Data structures
│   │   ├── core.go     # Game logic
│   │   ├── database.go # SQLite persistence
│   │   ├── llm.go      # AI integration
│   │   └── templates/  # HTML templates
│   └── cli/            # Standalone terminal client
├── scenarios/          # YAML scenario definitions
└── package.json        # Build scripts
```

## API Endpoints

### Core Endpoints
- `GET /health` - Server health check
- `GET /sessions` - List active sessions
- `POST /sessions` - Create new session
- `GET /sessions/:id` - Get session state

### Game Actions
- `POST /tools/get_state_summary` - Get game state summary
- `POST /tools/roll_check` - Perform dice rolls
- `POST /tools/apply_action` - Apply game actions

## Development

```bash
cd apps/dm-go
go run .                    # Development
go build -o dm-server .     # Production build
go test -v ./...            # Run tests
```

## Configuration

### Environment Variables
```bash
# Go Server
PORT=3000
DB_PATH=./dm-server.db
LLM_API_KEY=your-openai-key
LLM_MODEL=gpt-3.5-turbo
```

## Performance Benefits

- **5-10x faster** than Node.js equivalent
- **Lower memory usage** with Go's efficient GC
- **Single binary deployment** - no runtime dependencies
- **Native concurrency** with goroutines
- **Type safety** at compile time

## Migration from Legacy Version

This Go version replaces the complex TypeScript monorepo with:
- ✅ **Removed**: Node.js DM server and complex TypeScript packages
- ✅ **Kept**: Core game logic, web interface
- ✅ **Improved**: Performance, simplicity, maintainability

## Testing

Test the API directly:
```bash
# Health check
curl http://localhost:3000/health

# List sessions
curl http://localhost:3000/sessions
```

## Contributing

1. **Backend**: Focus on `apps/dm-go/`
2. **CLI**: Focus on `apps/cli/`
3. **Keep it simple**: Clean, minimal code

## License

MIT - Clean and simple, just like the codebase.
