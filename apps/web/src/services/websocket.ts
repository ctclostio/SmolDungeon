import { GameState, GameEvent, WSMessage } from '../types/game';

export type WebSocketEventHandler = (state: GameState) => void;
export type GameEventHandler = (event: GameEvent) => void;
export type ErrorHandler = (error: string) => void;

export class GameWebSocket {
  private ws: WebSocket | null = null;
  private sessionId: string;
  private onStateUpdate: WebSocketEventHandler | null = null;
  private onGameEvent: GameEventHandler | null = null;
  private onError: ErrorHandler | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(sessionId: string) {
    this.sessionId = sessionId;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // Use proxy path that Vite will redirect to Go backend
        const wsUrl = `ws://localhost:5173/ws/${this.sessionId}`;
        console.log('Connecting to WebSocket:', wsUrl);

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          reject(error);
        };

        this.ws.onclose = () => {
          console.log('WebSocket disconnected');
          this.handleReconnect();
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  private handleMessage(message: WSMessage) {
    switch (message.type) {
      case 'state_update':
        if (message.state && this.onStateUpdate) {
          this.onStateUpdate(message.state);
        }
        break;

      case 'event':
        if (message.event && this.onGameEvent) {
          this.onGameEvent(message.event);
        }
        break;

      case 'error':
        if (message.error && this.onError) {
          this.onError(message.error);
        }
        console.error('Game error:', message.error);
        break;

      default:
        console.warn('Unknown message type:', message);
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

      setTimeout(() => {
        this.connect().catch((error) => {
          console.error('Reconnection failed:', error);
        });
      }, delay);
    } else {
      console.error('Max reconnection attempts reached');
      if (this.onError) {
        this.onError('Connection lost. Please refresh the page.');
      }
    }
  }

  setStateUpdateHandler(handler: WebSocketEventHandler) {
    this.onStateUpdate = handler;
  }

  setGameEventHandler(handler: GameEventHandler) {
    this.onGameEvent = handler;
  }

  setErrorHandler(handler: ErrorHandler) {
    this.onError = handler;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}
