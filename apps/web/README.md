# SmolDungeon React Frontend

A modern, beautiful frontend for SmolDungeon built with React, TypeScript, and Phaser.

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type-safe JavaScript
- **Phaser 3** - 2D game rendering engine
- **Framer Motion** - Smooth animations
- **TailwindCSS** - Utility-first CSS framework
- **Zustand** - Lightweight state management
- **Vite** - Lightning-fast dev server and build tool

## Features

- ⚔️ **Tactical turn-based combat** with grid positioning
- ✨ **Smooth animations** using Framer Motion
- 🎮 **WebGL-powered game canvas** via Phaser 3
- 🎨 **Beautiful modern UI** with TailwindCSS
- ⚡ **Real-time updates** via WebSocket
- 🎯 **Type-safe** with TypeScript

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Go 1.21+ (for backend)

### Installation

```bash
# From the web directory
cd apps/web
npm install
```

### Development

```bash
# Start the frontend dev server (from project root)
npm run dev:web

# Or from the web directory
cd apps/web
npm run dev
```

The frontend will be available at `http://localhost:5173`

**Important:** You need the Go backend running on port 3000 for the frontend to work:

```bash
# In a separate terminal (from project root)
npm run dev

# Or
cd apps/dm-go
go run .
```

### Building for Production

```bash
# From project root
npm run build:web

# Or from web directory
cd apps/web
npm run build
```

This creates a `dist/` directory with optimized static files.

## Project Structure

```
apps/web/
├── src/
│   ├── components/          # React UI components
│   │   ├── ActionButtons.tsx
│   │   ├── CharacterCard.tsx
│   │   ├── GameUI.tsx
│   │   ├── HealthBar.tsx
│   │   └── PhaserGame.tsx
│   ├── game/                # Phaser game code
│   │   ├── scenes/
│   │   │   ├── CombatScene.ts
│   │   │   └── PreloadScene.ts
│   │   └── config.ts
│   ├── services/            # API and WebSocket clients
│   │   ├── api.ts
│   │   └── websocket.ts
│   ├── hooks/               # React hooks
│   │   └── useGameStore.ts
│   ├── types/               # TypeScript type definitions
│   │   └── game.ts
│   ├── styles/              # Global styles
│   │   └── index.css
│   ├── App.tsx              # Main app component
│   └── main.tsx             # Entry point
├── public/                  # Static assets
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

## How It Works

### Architecture

The frontend is a **client-side rendered (CSR)** single-page application that communicates with the Go backend via:

1. **REST API** - For game actions and state queries
2. **WebSocket** - For real-time game state updates

### Game Flow

1. User clicks "Start Demo" → Frontend calls `POST /sessions`
2. Backend creates a new game session and returns session ID + initial state
3. Frontend connects to WebSocket at `/ws/:sessionId`
4. User performs action → Frontend sends action via REST API
5. Backend processes action and broadcasts state update via WebSocket
6. Frontend updates Phaser scene and React UI

### State Management

- **Zustand store** manages global game state
- **React components** render UI based on state
- **Phaser scenes** render the game grid and characters
- **WebSocket** keeps state synchronized with backend

## API Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/sessions` | POST | Create new game session |
| `/sessions/:id/state` | GET | Get current game state |
| `/tools/apply_action` | POST | Apply game action |
| `/scenarios` | GET | List available scenarios |
| `/ws/:id` | WebSocket | Real-time game updates |

## Customization

### Changing Colors

Edit `tailwind.config.js`:

```js
colors: {
  combat: {
    player: '#10b981',  // Green for players
    enemy: '#ef4444',   // Red for enemies
  }
}
```

### Adding Animations

Framer Motion makes it easy:

```tsx
<motion.div
  initial={{ opacity: 0, y: 20 }}
  animate={{ opacity: 1, y: 0 }}
  transition={{ duration: 0.5 }}
>
  Your content
</motion.div>
```

### Modifying Game Grid

Edit `src/game/config.ts`:

```ts
export const GRID_SIZE = 5;  // Change grid dimensions
export const CELL_SIZE = 80; // Change cell size in pixels
```

## Troubleshooting

### Frontend can't connect to backend

- Ensure Go backend is running on port 3000
- Check CORS settings in `apps/dm-go/main.go`
- Verify Vite proxy settings in `vite.config.ts`

### WebSocket connection fails

- Check that the session ID is valid
- Verify the backend WebSocket endpoint is accessible
- Look for errors in browser console

### Build errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf node_modules/.vite
```

## Performance

- Phaser uses WebGL for hardware-accelerated rendering
- React components are optimized with proper memoization
- Vite provides instant hot module replacement (HMR)
- Production build is optimized and tree-shaken

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

Requires modern browser with WebGL support.

## Contributing

When adding new features:

1. Add TypeScript types in `src/types/game.ts`
2. Create reusable components in `src/components/`
3. Use TailwindCSS utility classes for styling
4. Add animations with Framer Motion
5. Keep Phaser code in `src/game/`

## License

MIT
