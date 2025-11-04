import { useState } from 'react';
import { motion } from 'framer-motion';
import { GameUI } from './components/GameUI';
import { gameApi } from './services/api';

type AppState = 'menu' | 'game';

function App() {
  const [appState, setAppState] = useState<AppState>('menu');
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const startDemo = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await gameApi.startSession();
      setSessionId(response.sessionId);
      setAppState('game');
    } catch (err) {
      console.error('Failed to start game:', err);
      setError(err instanceof Error ? err.message : 'Failed to start game');
    } finally {
      setIsLoading(false);
    }
  };

  const backToMenu = () => {
    setAppState('menu');
    setSessionId(null);
    setError(null);
  };

  if (appState === 'game' && sessionId) {
    return (
      <div>
        <GameUI sessionId={sessionId} />
        <motion.button
          className="fixed top-4 left-4 px-4 py-2 bg-slate-800 text-white rounded-lg hover:bg-slate-700 transition-colors"
          onClick={backToMenu}
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
        >
          ← Back to Menu
        </motion.button>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <motion.div
        className="max-w-2xl w-full glass p-12 rounded-2xl text-center"
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.5 }}
      >
        {/* Logo / Title */}
        <motion.h1
          className="text-6xl font-bold text-white mb-4 text-glow"
          initial={{ y: -20 }}
          animate={{ y: 0 }}
          transition={{ delay: 0.2 }}
        >
          ⚔️ SmolDungeon
        </motion.h1>

        <motion.p
          className="text-xl text-gray-300 mb-12"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.4 }}
        >
          Epic turn-based combat with stunning visuals
        </motion.p>

        {/* Error message */}
        {error && (
          <motion.div
            className="bg-red-500/20 border border-red-400 text-red-300 p-4 rounded-lg mb-6"
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
          >
            {error}
          </motion.div>
        )}

        {/* Buttons */}
        <div className="space-y-4">
          <motion.button
            className="w-full px-8 py-4 bg-gradient-to-r from-purple-600 to-blue-600 text-white text-xl font-bold rounded-lg hover:from-purple-500 hover:to-blue-500 transition-all transform hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
            onClick={startDemo}
            disabled={isLoading}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 }}
          >
            {isLoading ? (
              <span className="flex items-center justify-center gap-2">
                <motion.div
                  className="w-5 h-5 border-2 border-white border-t-transparent rounded-full"
                  animate={{ rotate: 360 }}
                  transition={{ duration: 1, repeat: Infinity, ease: 'linear' }}
                />
                Starting...
              </span>
            ) : (
              '🎮 Start Demo'
            )}
          </motion.button>

          <motion.button
            className="w-full px-8 py-4 bg-gradient-to-r from-green-600 to-teal-600 text-white text-xl font-bold rounded-lg hover:from-green-500 hover:to-teal-500 transition-all transform hover:scale-105"
            disabled={isLoading}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.7 }}
            onClick={() => alert('Scenario selection coming soon!')}
          >
            📜 Choose Scenario
          </motion.button>
        </div>

        {/* Features */}
        <motion.div
          className="mt-12 grid grid-cols-3 gap-6 text-gray-300"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.9 }}
        >
          <div>
            <div className="text-3xl mb-2">⚔️</div>
            <div className="text-sm font-semibold">Tactical Combat</div>
          </div>
          <div>
            <div className="text-3xl mb-2">✨</div>
            <div className="text-sm font-semibold">Smooth Animations</div>
          </div>
          <div>
            <div className="text-3xl mb-2">🎨</div>
            <div className="text-sm font-semibold">Beautiful UI</div>
          </div>
        </motion.div>

        {/* Footer */}
        <motion.p
          className="mt-12 text-sm text-gray-500"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1.1 }}
        >
          Powered by TypeScript + React + Phaser + Go
        </motion.p>
      </motion.div>
    </div>
  );
}

export default App;
