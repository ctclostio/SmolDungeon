import { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useGameStore } from '../hooks/useGameStore';
import { GameWebSocket } from '../services/websocket';
import { gameApi } from '../services/api';
import { PhaserGame } from './PhaserGame';
import { CharacterCard } from './CharacterCard';
import { ActionButtons } from './ActionButtons';
import { Character, Action, GameEvent } from '../types/game';

interface GameUIProps {
  sessionId: string;
}

export function GameUI({ sessionId }: GameUIProps) {
  const {
    gameState,
    selectedTarget,
    setGameState,
    setSelectedTarget,
    setError,
    getCurrentCharacter,
    getPlayerCharacters,
    getEnemyCharacters,
    isPlayerTurn,
  } = useGameStore();

  const [ws, setWs] = useState<GameWebSocket | null>(null);
  const [recentEvents, setRecentEvents] = useState<GameEvent[]>([]);
  const [isProcessing, setIsProcessing] = useState(false);

  // Initialize WebSocket connection
  useEffect(() => {
    const websocket = new GameWebSocket(sessionId);

    websocket.setStateUpdateHandler((state) => {
      console.log('State updated:', state);
      setGameState(state);
    });

    websocket.setGameEventHandler((event) => {
      console.log('Game event:', event);
      setRecentEvents((prev) => [...prev.slice(-9), event]);
    });

    websocket.setErrorHandler((error) => {
      console.error('WebSocket error:', error);
      setError(error);
    });

    websocket.connect().catch((error) => {
      console.error('Failed to connect WebSocket:', error);
      setError('Failed to connect to game server');
    });

    setWs(websocket);

    return () => {
      websocket.disconnect();
    };
  }, [sessionId, setGameState, setError]);

  // Load initial game state
  useEffect(() => {
    gameApi
      .getState(sessionId)
      .then((state) => {
        setGameState(state);
      })
      .catch((error) => {
        console.error('Failed to load game state:', error);
        setError('Failed to load game state');
      });
  }, [sessionId, setGameState, setError]);

  const handleAction = async (action: Action) => {
    if (isProcessing) return;

    try {
      setIsProcessing(true);
      const response = await gameApi.applyAction(sessionId, action);

      // Update state
      setGameState(response.state);

      // Clear selection
      setSelectedTarget(null);

      // Store events
      if (response.events) {
        setRecentEvents((prev) => [...prev, ...response.events]);
      }
    } catch (error) {
      console.error('Action failed:', error);
      setError(error instanceof Error ? error.message : 'Action failed');
    } finally {
      setIsProcessing(false);
    }
  };

  const currentCharacter = getCurrentCharacter();
  const playerCharacters = getPlayerCharacters();
  const enemyCharacters = getEnemyCharacters();
  const playerTurn = isPlayerTurn();

  if (!gameState) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-2xl text-white">Loading game...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <motion.div
          className="text-center mb-6"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <h1 className="text-4xl font-bold text-white mb-2">
            SmolDungeon
          </h1>
          <div className="text-lg text-gray-300">
            Round {gameState.round} •{' '}
            {currentCharacter && (
              <span className="text-yellow-400 font-semibold">
                {currentCharacter.name}'s Turn
              </span>
            )}
          </div>
        </motion.div>

        {/* Game complete message */}
        <AnimatePresence>
          {gameState.isComplete && (
            <motion.div
              className="glass p-6 rounded-lg mb-6 text-center"
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.9 }}
            >
              <h2 className="text-3xl font-bold text-white mb-2">
                {gameState.winner === 'player' && '🎉 Victory!'}
                {gameState.winner === 'enemy' && '💀 Defeat'}
                {gameState.winner === 'draw' && '🤝 Draw'}
              </h2>
              <p className="text-gray-300">The battle is over!</p>
            </motion.div>
          )}
        </AnimatePresence>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left sidebar - Player characters */}
          <div className="space-y-4">
            <h2 className="text-xl font-bold text-green-400 mb-4">Your Party</h2>
            {playerCharacters.map((character) => (
              <CharacterCard
                key={character.id}
                character={character}
                isActive={currentCharacter?.id === character.id}
              />
            ))}
          </div>

          {/* Center - Game canvas */}
          <div className="space-y-4">
            <PhaserGame gameState={gameState} events={recentEvents} />

            {/* Action buttons (only show on player turn) */}
            {playerTurn && currentCharacter && !gameState.isComplete && (
              <motion.div
                className="glass p-6 rounded-lg"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
              >
                <h3 className="text-lg font-bold text-white mb-4">
                  {currentCharacter.name}'s Actions
                </h3>
                <ActionButtons
                  character={currentCharacter}
                  selectedTarget={selectedTarget}
                  onAction={handleAction}
                  disabled={isProcessing || gameState.isComplete}
                />
              </motion.div>
            )}

            {/* AI turn indicator */}
            {!playerTurn && !gameState.isComplete && (
              <motion.div
                className="glass p-6 rounded-lg text-center"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
              >
                <motion.div
                  className="text-xl text-gray-300"
                  animate={{ opacity: [0.5, 1, 0.5] }}
                  transition={{ duration: 2, repeat: Infinity }}
                >
                  Enemy is thinking...
                </motion.div>
              </motion.div>
            )}
          </div>

          {/* Right sidebar - Enemy characters */}
          <div className="space-y-4">
            <h2 className="text-xl font-bold text-red-400 mb-4">Enemies</h2>
            {enemyCharacters.map((character) => (
              <CharacterCard
                key={character.id}
                character={character}
                isActive={currentCharacter?.id === character.id}
                isSelectable={playerTurn && !gameState.isComplete}
                isSelected={selectedTarget?.id === character.id}
                onClick={() => setSelectedTarget(character)}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
