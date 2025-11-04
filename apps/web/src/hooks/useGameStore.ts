import { create } from 'zustand';
import { GameState, Character, GameEvent, ID } from '../types/game';

interface GameStore {
  // State
  sessionId: string | null;
  gameState: GameState | null;
  selectedCharacter: Character | null;
  selectedTarget: Character | null;
  isLoading: boolean;
  error: string | null;
  narration: string | null;
  recentEvents: GameEvent[];

  // Actions
  setSessionId: (sessionId: string) => void;
  setGameState: (state: GameState) => void;
  setSelectedCharacter: (character: Character | null) => void;
  setSelectedTarget: (target: Character | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setNarration: (narration: string | null) => void;
  addEvent: (event: GameEvent) => void;
  clearEvents: () => void;
  reset: () => void;

  // Computed
  getCurrentCharacter: () => Character | null;
  getPlayerCharacters: () => Character[];
  getEnemyCharacters: () => Character[];
  isPlayerTurn: () => boolean;
}

export const useGameStore = create<GameStore>((set, get) => ({
  // Initial state
  sessionId: null,
  gameState: null,
  selectedCharacter: null,
  selectedTarget: null,
  isLoading: false,
  error: null,
  narration: null,
  recentEvents: [],

  // Actions
  setSessionId: (sessionId) => set({ sessionId }),

  setGameState: (gameState) => set({ gameState }),

  setSelectedCharacter: (character) => set({ selectedCharacter: character }),

  setSelectedTarget: (target) => set({ selectedTarget: target }),

  setLoading: (loading) => set({ isLoading: loading }),

  setError: (error) => set({ error }),

  setNarration: (narration) => set({ narration }),

  addEvent: (event) =>
    set((state) => ({
      recentEvents: [...state.recentEvents.slice(-9), event], // Keep last 10 events
    })),

  clearEvents: () => set({ recentEvents: [] }),

  reset: () =>
    set({
      sessionId: null,
      gameState: null,
      selectedCharacter: null,
      selectedTarget: null,
      isLoading: false,
      error: null,
      narration: null,
      recentEvents: [],
    }),

  // Computed values
  getCurrentCharacter: () => {
    const { gameState } = get();
    if (!gameState || gameState.turnOrder.length === 0) return null;

    const currentCharacterId = gameState.turnOrder[gameState.currentTurn];
    return gameState.characters.find((c) => c.id === currentCharacterId) || null;
  },

  getPlayerCharacters: () => {
    const { gameState } = get();
    if (!gameState) return [];
    return gameState.characters.filter((c) => c.isPlayer && c.stats.hp > 0);
  },

  getEnemyCharacters: () => {
    const { gameState } = get();
    if (!gameState) return [];
    return gameState.characters.filter((c) => !c.isPlayer && c.stats.hp > 0);
  },

  isPlayerTurn: () => {
    const currentChar = get().getCurrentCharacter();
    return currentChar?.isPlayer || false;
  },
}));
