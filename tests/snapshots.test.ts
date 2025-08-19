import { describe, it, expect } from 'vitest';
import { applyAction, createInitialState } from '@smol-dungeon/core';
import { LLMClient } from '@smol-dungeon/adapters';
import type { Character, Action } from '@smol-dungeon/schema';
import { randomUUID } from 'crypto';

const createTestCharacter = (isPlayer: boolean, name: string): Character => ({
  id: randomUUID(),
  name,
  stats: { hp: 30, maxHp: 30, attack: 5, defense: 3, speed: 4 },
  weapons: [{
    id: randomUUID(),
    name: 'Sword',
    damage: 6,
    accuracy: 85,
  }],
  abilities: [{
    id: randomUUID(),
    name: 'Power Strike',
    cooldown: 3,
    effect: 'damage' as const,
    power: 8,
  }],
  items: [],
  abilityCooldowns: {},
  isPlayer,
});

describe('Snapshot Tests', () => {
  const FIXED_SEED = 12345;

  it('should produce consistent combat narration for fixed seed', async () => {
    const player = createTestCharacter(true, 'Hero');
    const enemy = createTestCharacter(false, 'Goblin');
    const state = createInitialState([player], [enemy], FIXED_SEED);
    
    const attackAction: Action = {
      kind: 'Attack',
      attacker: player.id,
      target: enemy.id,
      weapon: player.weapons[0].id,
    };
    
    const resolution = applyAction(state, attackAction, FIXED_SEED);
    
    expect(resolution.logs).toMatchSnapshot('combat-logs-fixed-seed');
    expect(resolution.events).toMatchSnapshot('combat-events-fixed-seed');
    
    const stateSummary = {
      round: resolution.state.round,
      playerHp: resolution.state.characters.find(c => c.isPlayer)?.stats.hp,
      enemyHp: resolution.state.characters.find(c => !c.isPlayer)?.stats.hp,
      isComplete: resolution.state.isComplete,
    };
    
    expect(stateSummary).toMatchSnapshot('state-summary-fixed-seed');
  });

  it('should handle ability usage consistently', async () => {
    const player = createTestCharacter(true, 'Warrior');
    const enemy = createTestCharacter(false, 'Orc');
    const state = createInitialState([player], [enemy], FIXED_SEED);
    
    const abilityAction: Action = {
      kind: 'Ability',
      actor: player.id,
      ability: player.abilities[0].id,
      target: enemy.id,
    };
    
    const resolution = applyAction(state, abilityAction, FIXED_SEED);
    
    expect(resolution.logs).toMatchSnapshot('ability-logs-fixed-seed');
    expect(resolution.events).toMatchSnapshot('ability-events-fixed-seed');
    
    const cooldowns = resolution.state.characters
      .find(c => c.id === player.id)?.abilityCooldowns;
    
    expect(cooldowns).toMatchSnapshot('ability-cooldowns-fixed-seed');
  });

  it('should handle defend action consistently', async () => {
    const player = createTestCharacter(true, 'Guardian');
    const enemy = createTestCharacter(false, 'Skeleton');
    const state = createInitialState([player], [enemy], FIXED_SEED);
    
    const defendAction: Action = {
      kind: 'Defend',
      actor: player.id,
    };
    
    const resolution = applyAction(state, defendAction, FIXED_SEED);
    
    expect(resolution.logs).toMatchSnapshot('defend-logs-fixed-seed');
    expect(resolution.events).toMatchSnapshot('defend-events-fixed-seed');
    
    const playerStats = resolution.state.characters
      .find(c => c.id === player.id)?.stats;
    
    expect(playerStats).toMatchSnapshot('defend-stats-fixed-seed');
  });

  it('should handle complete combat sequence', async () => {
    const player = createTestCharacter(true, 'Adventurer');
    player.stats.attack = 10; // Make combat faster
    const enemy = createTestCharacter(false, 'Weak Goblin');
    enemy.stats.hp = 5; // Make enemy weaker
    enemy.stats.maxHp = 5;
    
    let state = createInitialState([player], [enemy], FIXED_SEED);
    const allLogs: string[] = [];
    const allEvents: any[] = [];
    
    let actionCount = 0;
    while (!state.isComplete && actionCount < 10) {
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
          
          const resolution = applyAction(state, attackAction, FIXED_SEED + actionCount);
          state = resolution.state;
          allLogs.push(...resolution.logs);
          allEvents.push(...resolution.events);
        }
      }
      
      actionCount++;
    }
    
    expect(allLogs).toMatchSnapshot('complete-combat-logs');
    expect(allEvents).toMatchSnapshot('complete-combat-events');
    expect({
      rounds: state.round,
      winner: state.winner,
      isComplete: state.isComplete,
      finalPlayerHp: state.characters.find(c => c.isPlayer)?.stats.hp,
      finalEnemyHp: state.characters.find(c => !c.isPlayer)?.stats.hp,
    }).toMatchSnapshot('complete-combat-result');
  });
});