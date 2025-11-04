import { motion } from 'framer-motion';
import { Character, Action } from '../types/game';

interface ActionButtonsProps {
  character: Character;
  selectedTarget: Character | null;
  onAction: (action: Action) => void;
  disabled?: boolean;
}

export function ActionButtons({
  character,
  selectedTarget,
  onAction,
  disabled = false,
}: ActionButtonsProps) {
  const handleAttack = () => {
    if (!selectedTarget || !character.weapons[0]) return;

    onAction({
      kind: 'attack',
      attacker: character.id,
      target: selectedTarget.id,
      weapon: character.weapons[0].id,
    });
  };

  const handleDefend = () => {
    onAction({
      kind: 'defend',
      actor: character.id,
    });
  };

  const handleAbility = (abilityId: string) => {
    if (!selectedTarget) return;

    onAction({
      kind: 'ability',
      actor: character.id,
      ability: abilityId,
      target: selectedTarget.id,
    });
  };

  const handleItem = (itemId: string) => {
    onAction({
      kind: 'item',
      actor: character.id,
      item: itemId,
    });
  };

  const handleFlee = () => {
    onAction({
      kind: 'flee',
      actor: character.id,
    });
  };

  const container = {
    hidden: { opacity: 0 },
    show: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
      },
    },
  };

  const item = {
    hidden: { opacity: 0, y: 20 },
    show: { opacity: 1, y: 0 },
  };

  return (
    <motion.div
      className="space-y-4"
      variants={container}
      initial="hidden"
      animate="show"
    >
      {/* Main actions */}
      <div className="grid grid-cols-2 gap-3">
        <motion.button
          variants={item}
          className="action-btn attack"
          onClick={handleAttack}
          disabled={disabled || !selectedTarget || !character.weapons[0]}
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
        >
          ⚔️ Attack
          {selectedTarget && <div className="text-xs opacity-80">{selectedTarget.name}</div>}
        </motion.button>

        <motion.button
          variants={item}
          className="action-btn defend"
          onClick={handleDefend}
          disabled={disabled}
          whileHover={{ scale: 1.05 }}
          whileTap={{ scale: 0.95 }}
        >
          🛡️ Defend
        </motion.button>
      </div>

      {/* Abilities */}
      {character.abilities.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm font-semibold text-gray-300">Abilities</div>
          <div className="grid grid-cols-2 gap-2">
            {character.abilities.map((ability) => {
              const cooldown = character.abilityCooldowns[ability.id] || 0;
              const isOnCooldown = cooldown > 0;

              return (
                <motion.button
                  key={ability.id}
                  variants={item}
                  className="action-btn ability text-sm"
                  onClick={() => handleAbility(ability.id)}
                  disabled={disabled || isOnCooldown || !selectedTarget}
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                >
                  ✨ {ability.name}
                  {isOnCooldown && (
                    <div className="text-xs opacity-80">Cooldown: {cooldown}</div>
                  )}
                </motion.button>
              );
            })}
          </div>
        </div>
      )}

      {/* Items */}
      {character.items.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm font-semibold text-gray-300">Items</div>
          <div className="grid grid-cols-2 gap-2">
            {character.items.map((item) => (
              <motion.button
                key={item.id}
                variants={item}
                className="action-btn item text-sm"
                onClick={() => handleItem(item.id)}
                disabled={disabled}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
              >
                💊 {item.name}
              </motion.button>
            ))}
          </div>
        </div>
      )}

      {/* Flee */}
      <motion.button
        variants={item}
        className="action-btn flee w-full text-sm"
        onClick={handleFlee}
        disabled={disabled}
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.95 }}
      >
        🏃 Flee
      </motion.button>

      {/* Help text */}
      {!selectedTarget && character.weapons.length > 0 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="text-sm text-yellow-400 text-center bg-yellow-500/10 p-2 rounded"
        >
          👆 Select an enemy to attack
        </motion.div>
      )}
    </motion.div>
  );
}
