import type { Character, State, Id } from '@smol-dungeon/schema';
import { SeededRNG } from './rng.js';

export function createInitialState(players: Character[], enemies: Character[], seed: number): State {
  const rng = new SeededRNG(seed);
  const allCharacters = [...players, ...enemies];
  
  const turnOrder = allCharacters
    .map(char => ({ id: char.id, initiative: char.stats.speed + rng.rollD20() }))
    .sort((a, b) => b.initiative - a.initiative)
    .map(entry => entry.id);

  return {
    round: 1,
    characters: allCharacters,
    turnOrder,
    currentTurn: 0,
    isComplete: false,
  };
}

export function getCurrentCharacter(state: State): Character | undefined {
  const currentId = state.turnOrder[state.currentTurn];
  return state.characters.find(c => c.id === currentId);
}

export function getCharacterById(state: State, id: Id): Character | undefined {
  return state.characters.find(c => c.id === id);
}

export function getStateSummary(state: State): string {
  const players = state.characters.filter(c => c.isPlayer);
  const enemies = state.characters.filter(c => !c.isPlayer);
  
  let summary = `Round ${state.round}\n\n`;
  
  summary += "Players:\n";
  players.forEach(p => {
    const status = p.stats.hp > 0 ? `${p.stats.hp}/${p.stats.maxHp} HP` : 'DEFEATED';
    summary += `  ${p.name}: ${status}\n`;
  });
  
  summary += "\nEnemies:\n";
  enemies.forEach(e => {
    const status = e.stats.hp > 0 ? `${e.stats.hp}/${e.stats.maxHp} HP` : 'DEFEATED';
    summary += `  ${e.name}: ${status}\n`;
  });
  
  if (state.isComplete) {
    summary += `\nCombat Complete! Winner: ${state.winner || 'Draw'}`;
  } else {
    const currentChar = getCurrentCharacter(state);
    summary += `\nCurrent Turn: ${currentChar?.name || 'Unknown'}`;
  }
  
  return summary;
}