# How to Play SmolDungeon

SmolDungeon now supports **THREE different ways to play**! Choose your preferred experience:

---

## 🎮 Option 1: Modern Web Frontend (Recommended)

**The SNAZZY, PRETTY, and EASY experience!**

### Features:
- ✨ WebGL-rendered game grid (60 FPS)
- 💥 Attack animations and particle effects
- 📊 Animated health bars
- 🎨 Beautiful card-based UI
- 🎯 Click to select targets
- ⚡ Real-time WebSocket updates

### How to Play:

**Terminal 1 - Backend:**
```bash
cd apps/dm-go
go run . demo
```

**Terminal 2 - Frontend:**
```bash
cd apps/web
npm run dev
```

**Open your browser:** http://localhost:5173

---

## 💻 Option 2: Terminal/CLI (NEW!)

**For terminal enthusiasts and command-line warriors!**

### Features:
- 🎨 Colored terminal output
- 📊 ASCII health bars
- ⚔️  Interactive combat
- 🖥️  Works anywhere with a terminal
- 🤖 Can be automated/scripted

### How to Play:

**Terminal 1 - Start backend:**
```bash
cd apps/dm-go
go run . demo
```

**Terminal 2 - Play the game:**
```bash
cd apps/cli
go run .
```

**Or use the shortcut:**
```bash
npm run play
```

### Controls:
```
1 - ⚔️  Attack (then select target)
2 - 🛡️  Defend
3 - 🏃 Flee
```

---

## 🌐 Option 3: Classic Web UI

**The original server-side rendered experience**

### Features:
- 🔙 Classic HTML interface
- 🎲 Full game functionality
- 📱 Works on any browser
- 🚀 No build step required

### How to Play:

```bash
cd apps/dm-go
go run . demo
```

**Open your browser:** http://localhost:3000

---

## 🎯 Quick Comparison

| Feature | Modern Web | CLI | Classic Web |
|---------|-----------|-----|-------------|
| **Visual Effects** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐ |
| **Animations** | ⭐⭐⭐⭐⭐ | ⭐ | ⭐ |
| **Terminal-Friendly** | ❌ | ⭐⭐⭐⭐⭐ | ❌ |
| **Mobile Support** | ⭐⭐⭐⭐ | ❌ | ⭐⭐⭐ |
| **Build Required** | Yes (npm) | No | No |
| **Speed** | ⚡⚡⚡⚡⚡ | ⚡⚡⚡⚡ | ⚡⚡⚡ |
| **Coolness Factor** | 🔥🔥🔥🔥🔥 | 🔥🔥🔥🔥 | 🔥🔥 |

---

## 🛠️ Development Commands

From the project root:

```bash
# Backend
npm run dev              # Start Go backend (port 3000)

# Frontends
npm run dev:web          # Start React frontend (port 5173)
npm run play             # Play CLI version

# Building
npm run build            # Build Go backend binary
npm run build:web        # Build React frontend for production
npm run build:cli        # Build CLI client binary
```

---

## 🎮 Gameplay Tips

### Combat Strategy:
1. **Attack** - Use your best weapon to deal damage
2. **Defend** - Boost defense for one turn (AI uses this a lot!)
3. **Abilities** - Special moves with cooldowns
4. **Items** - Consumables for healing/buffs
5. **Flee** - Run away (might fail!)

### Target Selection:
- Focus fire on one enemy to eliminate threats
- Check enemy HP before attacking
- Low HP enemies are easier to finish off

### Turn Order:
- Based on Speed stat + d20 roll
- Watch the turn indicator to know who's next
- Plan ahead!

---

## 🎨 CLI Output Example

```
╔═══════════════════════════════════════════════╗
║           ⚔️  SMOLDUNGEON CLI ⚔️             ║
╚═══════════════════════════════════════════════╝

═══════════════════════════════════════════════
  Round 1
═══════════════════════════════════════════════

🗡️  YOUR PARTY:
  Fighter        HP:  30/30 [████████████████████]
     ATK: 6  DEF: 4  SPD: 3  Pos: (0,0)

👹 ENEMIES:
  Goblin Warrior   HP:  15/15 [████████████████████]
     ATK: 4  DEF: 2  SPD: 5  Pos: (1,1)
  Goblin Archer    HP:  12/12 [████████████████████]
     ATK: 5  DEF: 1  SPD: 6  Pos: (-1,2)

⚔️  Fighter's Turn!
What will you do?
1. ⚔️  Attack
2. 🛡️  Defend
3. 🏃 Flee
```

---

## 🎊 Recommended Experience

For the **BEST** experience:
1. **New players** → Modern Web Frontend (most visual feedback)
2. **Terminal lovers** → CLI (works great in tmux/screen)
3. **Quick testing** → CLI (fastest to start)
4. **Remote servers** → CLI (SSH-friendly)
5. **Showing off** → Modern Web Frontend (impress your friends!)

---

## 🐛 Troubleshooting

### "Connection refused"
Make sure the backend is running on port 3000:
```bash
cd apps/dm-go && go run . demo
```

### "npm: command not found"
Install Node.js 18+ from https://nodejs.org
(Only needed for Modern Web Frontend)

### CLI colors not showing
Your terminal might not support ANSI colors. Try:
- Windows: Use Windows Terminal or PowerShell
- macOS/Linux: Most terminals work by default

---

## 🚀 What's Next?

Try all three versions and see which one you prefer! Each offers a different experience:
- **Modern Web** = Professional game feel
- **CLI** = Hackery retro vibes
- **Classic Web** = Simple and reliable

**Enjoy your dungeon crawling adventure!** ⚔️🐉

Made with ❤️ using Go, React, TypeScript, Phaser, and pure determination!
