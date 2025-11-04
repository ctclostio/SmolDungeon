import { motion } from 'framer-motion';
import { Character } from '../types/game';
import { HealthBar } from './HealthBar';

interface CharacterCardProps {
  character: Character;
  isActive?: boolean;
  isSelectable?: boolean;
  isSelected?: boolean;
  onClick?: () => void;
}

export function CharacterCard({
  character,
  isActive = false,
  isSelectable = false,
  isSelected = false,
  onClick,
}: CharacterCardProps) {
  const cardClass = `
    character-card
    ${character.isPlayer ? 'player' : 'enemy'}
    ${isActive ? 'active' : ''}
    ${isSelectable ? 'cursor-pointer hover:scale-105' : ''}
    ${isSelected ? 'ring-4 ring-yellow-400' : ''}
  `;

  return (
    <motion.div
      className={cardClass}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.8 }}
      whileHover={isSelectable ? { scale: 1.05 } : {}}
      whileTap={isSelectable ? { scale: 0.95 } : {}}
      onClick={onClick}
    >
      {/* Character name and status */}
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-lg font-bold text-white">{character.name}</h3>
        {isActive && (
          <motion.div
            className="text-xs bg-blue-500 text-white px-2 py-1 rounded-full font-semibold"
            animate={{ scale: [1, 1.1, 1] }}
            transition={{ duration: 1, repeat: Infinity }}
          >
            ACTIVE
          </motion.div>
        )}
      </div>

      {/* Health bar */}
      <HealthBar stats={character.stats} className="mb-3" />

      {/* Stats */}
      <div className="grid grid-cols-3 gap-2 text-xs text-gray-300">
        <div className="text-center">
          <div className="font-semibold text-red-400">⚔️ {character.stats.attack}</div>
          <div className="text-gray-500">ATK</div>
        </div>
        <div className="text-center">
          <div className="font-semibold text-blue-400">🛡️ {character.stats.defense}</div>
          <div className="text-gray-500">DEF</div>
        </div>
        <div className="text-center">
          <div className="font-semibold text-yellow-400">⚡ {character.stats.speed}</div>
          <div className="text-gray-500">SPD</div>
        </div>
      </div>

      {/* Position */}
      <div className="mt-2 text-xs text-gray-500 text-center">
        Position: ({character.position.x}, {character.position.y})
      </div>

      {/* Abilities on cooldown */}
      {Object.keys(character.abilityCooldowns).length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {Object.entries(character.abilityCooldowns).map(([abilityId, turns]) => {
            const ability = character.abilities.find((a) => a.id === abilityId);
            if (!ability || turns === 0) return null;
            return (
              <span
                key={abilityId}
                className="text-xs bg-purple-500/20 text-purple-300 px-2 py-0.5 rounded"
              >
                {ability.name}: {turns} turn{turns > 1 ? 's' : ''}
              </span>
            );
          })}
        </div>
      )}
    </motion.div>
  );
}
