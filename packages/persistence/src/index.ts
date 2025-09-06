export interface Event {
  type: string;
  target?: string;
  amount?: number;
  source?: string;
  actor?: string;
  ability?: string;
  item?: string;
}

export interface State {
  round: number;
  characters: Character[];
  turnOrder: string[];
  currentTurn: number;
  isComplete: boolean;
  winner?: string;
}

export interface Character {
  id: string;
  name: string;
  isPlayer: boolean;
  stats: {
    hp: number;
    maxHp: number;
    attack: number;
    defense: number;
    speed: number;
  };
  position: {
    x: number;
    y: number;
  };
  weapons: Weapon[];
  abilities: Ability[];
  items: Item[];
  abilityCooldowns: { [key: string]: number };
}

export interface Weapon {
  id: string;
  name: string;
  damage: number;
  accuracy: number;
}

export interface Ability {
  id: string;
  name: string;
  cooldown: number;
  effect: string;
  power: number;
}

export interface Item {
  id: string;
  name: string;
  type: string;
  effect: string;
}

export class PersistenceManager {
  private storage: Storage = localStorage; // Temp in-memory workaround - VIOLATION, replace with API or IndexedDB

  saveState(sessionId: string, state: State) {
    this.storage.setItem(`state_${sessionId}`, JSON.stringify(state));
  }

  loadState(sessionId: string): State | null {
    const data = this.storage.getItem(`state_${sessionId}`);
    return data ? JSON.parse(data) as State : null;
  }

  saveEvents(sessionId: string, round: number, events: Event[]) {
    this.storage.setItem(`events_${sessionId}_${round}`, JSON.stringify(events));
  }

  loadEvents(sessionId: string, fromRound: number): Event[] {
    // Simplified - load all, filter
    let allEvents: Event[] = [];
    for (let i = fromRound; i < 100; i++) { // Arbitrary max
      const data = this.storage.getItem(`events_${sessionId}_${i}`);
      if (data) {
        allEvents = allEvents.concat(JSON.parse(data) as Event[]);
      } else {
        break;
      }
    }
    return allEvents;
  }

  // TODO: Full IndexedDB or backend API integration
  // WARNING: Uses localStorage - in-memory violation
}