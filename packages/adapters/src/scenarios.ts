import YAML from 'js-yaml';
import type { Character } from '@smol-dungeon/schema';
import { randomUUID } from 'crypto';

export interface Scenario {
  name: string;
  description: string;
  players: CharacterTemplate[];
  enemies: CharacterTemplate[];
  context: string;
}

export interface CharacterTemplate {
  name: string;
  stats: {
    hp: number;
    maxHp: number;
    attack: number;
    defense: number;
    speed: number;
  };
  weapons: Array<{
    name: string;
    damage: number;
    accuracy: number;
  }>;
  abilities: Array<{
    name: string;
    cooldown: number;
    effect: 'damage' | 'heal' | 'buff' | 'debuff';
    power: number;
  }>;
  items: Array<{
    name: string;
    type: 'consumable' | 'equipment';
    effect: string;
  }>;
}

export function parseScenario(yamlContent: string): Scenario {
  const data = YAML.load(yamlContent) as any;
  
  return {
    name: data.name || 'Unknown Scenario',
    description: data.description || '',
    players: data.players || [],
    enemies: data.enemies || [],
    context: data.context || '',
  };
}

export function createCharacterFromTemplate(template: CharacterTemplate, isPlayer: boolean): Character {
  return {
    id: randomUUID(),
    name: template.name,
    stats: template.stats,
    weapons: template.weapons.map(w => ({ ...w, id: randomUUID() })),
    abilities: template.abilities.map(a => ({ ...a, id: randomUUID() })),
    items: template.items.map(i => ({ ...i, id: randomUUID() })),
    abilityCooldowns: {},
    isPlayer,
  };
}

export function getDefaultScenario(): Scenario {
  return {
    name: 'Goblin Ambush',
    description: 'A group of goblins attacks the party on a forest path.',
    context: 'The party is traveling through a dark forest when goblins leap from the bushes!',
    players: [
      {
        name: 'Fighter',
        stats: { hp: 30, maxHp: 30, attack: 6, defense: 4, speed: 3 },
        weapons: [
          { name: 'Sword', damage: 8, accuracy: 85 },
          { name: 'Shield Bash', damage: 4, accuracy: 90 },
        ],
        abilities: [
          { name: 'Power Attack', cooldown: 3, effect: 'damage', power: 12 },
          { name: 'Second Wind', cooldown: 5, effect: 'heal', power: 15 },
        ],
        items: [
          { name: 'Health Potion', type: 'consumable', effect: 'heal 20 HP' },
        ],
      },
    ],
    enemies: [
      {
        name: 'Goblin Warrior',
        stats: { hp: 15, maxHp: 15, attack: 4, defense: 2, speed: 5 },
        weapons: [
          { name: 'Rusty Sword', damage: 5, accuracy: 75 },
        ],
        abilities: [
          { name: 'Sneaky Strike', cooldown: 4, effect: 'damage', power: 8 },
        ],
        items: [],
      },
      {
        name: 'Goblin Archer',
        stats: { hp: 12, maxHp: 12, attack: 5, defense: 1, speed: 6 },
        weapons: [
          { name: 'Short Bow', damage: 6, accuracy: 80 },
        ],
        abilities: [
          { name: 'Aimed Shot', cooldown: 3, effect: 'damage', power: 10 },
        ],
        items: [],
      },
    ],
  };
}