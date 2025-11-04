import { useEffect, useRef } from 'react';
import Phaser from 'phaser';
import { gameConfig } from '../game/config';
import { GameState, GameEvent } from '../types/game';

interface PhaserGameProps {
  gameState: GameState | null;
  events?: GameEvent[];
}

export function PhaserGame({ gameState, events = [] }: PhaserGameProps) {
  const gameRef = useRef<Phaser.Game | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!containerRef.current || gameRef.current) return;

    // Create Phaser game instance
    gameRef.current = new Phaser.Game({
      ...gameConfig,
      parent: containerRef.current,
    });

    return () => {
      // Cleanup on unmount
      if (gameRef.current) {
        gameRef.current.destroy(true);
        gameRef.current = null;
      }
    };
  }, []);

  // Update game state when it changes
  useEffect(() => {
    if (!gameRef.current || !gameState) return;

    // Emit state update event to Phaser
    gameRef.current.events.emit('stateUpdate', gameState);
  }, [gameState]);

  // Process game events
  useEffect(() => {
    if (!gameRef.current || events.length === 0) return;

    const game = gameRef.current;

    events.forEach((event) => {
      switch (event.type) {
        case 'damage':
          if (event.target && event.amount) {
            game.events.emit('characterDamaged', {
              characterId: event.target,
              amount: event.amount,
            });
          }
          break;

        case 'heal':
          if (event.target && event.amount) {
            game.events.emit('characterHealed', {
              characterId: event.target,
              amount: event.amount,
            });
          }
          break;

        case 'miss':
          if (event.target) {
            game.events.emit('characterMissed', event.target);
          }
          break;

        case 'attack':
          if (event.source && event.target) {
            game.events.emit('attackAnimation', {
              attackerId: event.source,
              targetId: event.target,
            });
          }
          break;

        default:
          break;
      }
    });
  }, [events]);

  return (
    <div className="game-canvas-container">
      <div ref={containerRef} id="phaser-game" />
    </div>
  );
}
