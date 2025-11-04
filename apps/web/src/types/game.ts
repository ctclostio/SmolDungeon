/**
 * TypeScript types matching the Go backend structures
 */

export type ID = string;

export interface Stat {
  hp: number;
  maxHp: number;
  attack: number;
  defense: number;
  speed: number;
}

export interface Position {
  x: number;
  y: number;
}

export interface Weapon {
  id: ID;
  name: string;
  damage: number;
  accuracy: number;
}

export interface Ability {
  id: ID;
  name: string;
  cooldown: number;
  effect: 'damage' | 'heal' | 'buff' | 'debuff';
  power: number;
}

export interface Item {
  id: ID;
  name: string;
  type: 'consumable' | 'equipment';
  effect: string;
}

export interface Character {
  id: ID;
  name: string;
  stats: Stat;
  position: Position;
  weapons: Weapon[];
  abilities: Ability[];
  items: Item[];
  abilityCooldowns: Record<string, number>;
  isPlayer: boolean;
}

export type ActionKind =
  | 'attack'
  | 'defend'
  | 'ability'
  | 'item'
  | 'flee'
  | 'initiative';

export interface Action {
  kind: ActionKind;
  attacker?: ID;
  target?: ID;
  weapon?: ID;
  actor?: ID;
  ability?: ID;
  item?: ID;
}

export type EventType =
  | 'damage'
  | 'heal'
  | 'attack'
  | 'miss'
  | 'defend'
  | 'ability_used'
  | 'item_used'
  | 'death'
  | 'flee'
  | 'turn_start'
  | 'turn_end'
  | 'round_start'
  | 'game_start'
  | 'game_end';

export interface GameEvent {
  id?: string;
  type: EventType;
  target?: ID;
  amount?: number;
  source?: ID;
  actor?: ID;
  ability?: ID;
  item?: ID;
}

export interface GameState {
  round: number;
  characters: Character[];
  turnOrder: ID[];
  currentTurn: number;
  isComplete: boolean;
  winner?: 'player' | 'enemy' | 'draw';
}

/**
 * WebSocket message types
 */
export interface WSMessage {
  type: 'state_update' | 'event' | 'error';
  state?: GameState;
  event?: GameEvent;
  error?: string;
}

/**
 * API response types
 */
export interface ActionResponse {
  state: GameState;
  events: GameEvent[];
  narration?: string;
}

export interface Scenario {
  name: string;
  description: string;
  players: Character[];
  enemies: Character[];
}
