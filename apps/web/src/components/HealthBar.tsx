import { motion } from 'framer-motion';
import { Stat } from '../types/game';

interface HealthBarProps {
  stats: Stat;
  className?: string;
}

export function HealthBar({ stats, className = '' }: HealthBarProps) {
  const healthPercent = (stats.hp / stats.maxHp) * 100;

  const getHealthColor = () => {
    if (healthPercent > 60) return 'high';
    if (healthPercent > 30) return 'medium';
    return 'low';
  };

  return (
    <div className={`space-y-1 ${className}`}>
      <div className="flex justify-between items-center text-xs text-gray-300">
        <span className="font-semibold">HP</span>
        <span>
          {stats.hp} / {stats.maxHp}
        </span>
      </div>
      <div className="health-bar">
        <motion.div
          className={`health-bar-fill ${getHealthColor()}`}
          initial={{ width: 0 }}
          animate={{ width: `${healthPercent}%` }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
        />
      </div>
    </div>
  );
}
