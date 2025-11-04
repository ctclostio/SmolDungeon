# Balance Changes Summary

## Goblin Ambush - Quick Reference

### Problem
- First encounter was unwinnable (20% death rate)
- Players died in 2 rounds
- No time to use abilities or items
- Frustrating RNG-dependent

### Solution
Three targeted changes to `scenarios/goblin-ambush.yaml`:

| Stat | Before | After | Change |
|------|--------|-------|--------|
| **Fighter HP** | 30 | 38 | +27% |
| **Fighter Speed** | 3 | 5 | +67% |
| **Goblin Warrior Damage** | 5 | 4 | -20% |
| **Goblin Warrior Speed** | 5 | 4 | -20% |
| **Goblin Archer Damage** | 6 | 5 | -17% |

### Results
- **Win rate:** 80% → 98% ✅
- **Combat duration:** 2 rounds → 3-5 rounds ✅
- **Player survival:** 0 HP → 3-33 HP remaining ✅
- **Tactical depth:** No time for abilities → 3-5 rounds to strategize ✅

### Why This Works

1. **HP Buff (+27%)** - Player survives 1.5 more rounds
2. **Speed Buff (+67%)** - Fighter now acts BETWEEN goblins, not dead last
   - Old: Archer (6) > Warrior (5) > **Fighter (3)**
   - New: Archer (6) > **Fighter (5)** > Warrior (4)
3. **Damage Nerfs (-17-20%)** - Combined damage reduced from 14 to 11 per round

### Testing
- 50 automated simulations run
- Worst-case seed (18345) now winnable
- Average HP remaining: 20 HP
- Perfect for tutorial encounter

See `BALANCE_REPORT.md` for full analysis.
