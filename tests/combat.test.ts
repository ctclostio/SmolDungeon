import { describe, it, expect } from 'vitest';
import fc from 'fast-check';
import { applyAction, createInitialState, checkCombatEnd } from '@smol-dungeon/core';
import type { Character, Action, State } from '@smol-dungeon/schema';
import { randomUUID } from 'crypto';

const createTestCharacter = (isPlayer: boolean, name: string): Character => ({
  id: randomUUID(),
  name,
  stats: { hp: 30, maxHp: 30, attack: 5, defense: 3, speed: 4 },
  weapons: [{
    id: randomUUID(),
    name: 'Test Weapon',
    damage: 6,
    accuracy: 85,
  }],
  abilities: [{
    id: randomUUID(),
    name: 'Test Ability',
    cooldown: 3,
    effect: 'damage' as const,
    power: 8,
  }],
  items: [{
    id: randomUUID(),
    name: 'Health Potion',
    type: 'consumable' as const,
    effect: 'heal 20 HP',
  }],
  abilityCooldowns: {},
  isPlayer,
});

describe('Combat Engine Properties', () => {
  it('should never produce negative damage', () => {
    fc.assert(fc.property(
      fc.integer({ min: 1, max: 1000 }), // seed
      (seed) => {
        const player = createTestCharacter(true, 'Player');
        const enemy = createTestCharacter(false, 'Enemy');
        const state = createInitialState([player], [enemy], seed);
        
        const attackAction: Action = {
          kind: 'Attack',
          attacker: player.id,
          target: enemy.id,
          weapon: player.weapons[0].id,
        };
        
        const resolution = applyAction(state, attackAction, seed);
        
        const damageEvents = resolution.events.filter(e => e.type === 'damage');
        damageEvents.forEach(event => {
          expect(event.amount).toBeGreaterThanOrEqual(0);
        });
        
        resolution.state.characters.forEach(char => {
          expect(char.stats.hp).toBeGreaterThanOrEqual(0);
          expect(char.stats.hp).toBeLessThanOrEqual(char.stats.maxHp);
        });
      }
    ));
  });

  it('should terminate combat within reasonable rounds', () => {
    fc.assert(fc.property(
      fc.integer({ min: 1, max: 1000 }),
      (seed) => {
        const player = createTestCharacter(true, 'Player');
        const enemy = createTestCharacter(false, 'Enemy');
        let state = createInitialState([player], [enemy], seed);
        
        let rounds = 0;
        const maxRounds = 50;
        
        while (!checkCombatEnd(state) && rounds < maxRounds) {
          const currentChar = state.characters.find(c => 
            c.id === state.turnOrder[state.currentTurn]
          );
          
          if (currentChar) {
            const targets = state.characters.filter(c => 
              c.isPlayer !== currentChar.isPlayer && c.stats.hp > 0
            );
            
            if (targets.length > 0) {
              const attackAction: Action = {
                kind: 'Attack',
                attacker: currentChar.id,
                target: targets[0].id,
                weapon: currentChar.weapons[0].id,
              };
              
              const resolution = applyAction(state, attackAction, seed + rounds);
              state = resolution.state;
            }
          }
          
          rounds++;
        }
        
        expect(rounds).toBeLessThan(maxRounds);
      }
    ));
  });

  it('should be deterministic with same seed', () => {
    fc.assert(fc.property(
      fc.integer({ min: 1, max: 1000 }),
      (seed) => {
        const player1 = createTestCharacter(true, 'Player');
        const enemy1 = createTestCharacter(false, 'Enemy');
        const state1 = createInitialState([player1], [enemy1], seed);
        
        const player2 = createTestCharacter(true, 'Player');
        const enemy2 = createTestCharacter(false, 'Enemy');
        const state2 = createInitialState([player2], [enemy2], seed);
        
        const attackAction1: Action = {
          kind: 'Attack',
          attacker: player1.id,
          target: enemy1.id,
          weapon: player1.weapons[0].id,
        };
        
        const attackAction2: Action = {
          kind: 'Attack',
          attacker: player2.id,
          target: enemy2.id,
          weapon: player2.weapons[0].id,
        };
        
        const resolution1 = applyAction(state1, attackAction1, seed);
        const resolution2 = applyAction(state2, attackAction2, seed);
        
        expect(resolution1.events.length).toBe(resolution2.events.length);
        expect(resolution1.logs.length).toBe(resolution2.logs.length);
        
        resolution1.events.forEach((event, index) => {
          const event2 = resolution2.events[index];
          expect(event.type).toBe(event2.type);
          if (event.type === 'damage' && event2.type === 'damage') {
            expect(event.amount).toBe(event2.amount);
          }
        });
      }
    ));
  });

  it('should maintain valid turn order', () => {
    fc.assert(fc.property(
      fc.integer({ min: 1, max: 1000 }),
      (seed) => {
        const player = createTestCharacter(true, 'Player');
        const enemy = createTestCharacter(false, 'Enemy');
        const state = createInitialState([player], [enemy], seed);
        
        expect(state.turnOrder).toHaveLength(2);
        expect(state.turnOrder).toContain(player.id);
        expect(state.turnOrder).toContain(enemy.id);
        expect(state.currentTurn).toBeGreaterThanOrEqual(0);
        expect(state.currentTurn).toBeLessThan(state.turnOrder.length);
      }
    ));
  });
});