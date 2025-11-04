import { Action, ActionResponse, GameState, Scenario } from '../types/game';

const API_BASE_URL = '/api';

class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: any
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new ApiError(
        errorData.error || `HTTP ${response.status}: ${response.statusText}`,
        response.status,
        errorData
      );
    }

    return await response.json();
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError(
      error instanceof Error ? error.message : 'Network error',
      0
    );
  }
}

export const gameApi = {
  /**
   * Apply an action to the game
   */
  async applyAction(sessionId: string, action: Action): Promise<ActionResponse> {
    return fetchApi<ActionResponse>('/tools/apply_action', {
      method: 'POST',
      body: JSON.stringify({
        sessionId,
        action,
      }),
    });
  },

  /**
   * Get current game state
   */
  async getState(sessionId: string): Promise<GameState> {
    return fetchApi<GameState>(`/sessions/${sessionId}/state`);
  },

  /**
   * Start a new game session
   */
  async startSession(scenarioName?: string): Promise<{ sessionId: string; state: GameState }> {
    return fetchApi<{ sessionId: string; state: GameState }>('/sessions', {
      method: 'POST',
      body: JSON.stringify({ scenarioName }),
    });
  },

  /**
   * Get available scenarios
   */
  async getScenarios(): Promise<Scenario[]> {
    return fetchApi<Scenario[]>('/scenarios');
  },

  /**
   * Get state summary with narration
   */
  async getStateSummary(sessionId: string): Promise<{ summary: string; narration?: string }> {
    return fetchApi<{ summary: string; narration?: string }>(
      `/tools/get_state_summary?sessionId=${sessionId}`
    );
  },

  /**
   * Generate narration for current state
   */
  async generateNarration(sessionId: string, prompt: string): Promise<{ narration: string }> {
    return fetchApi<{ narration: string }>('/llm/generate_narration', {
      method: 'POST',
      body: JSON.stringify({
        sessionId,
        prompt,
      }),
    });
  },
};

export { ApiError };
