import Phaser from 'phaser';
import { CombatScene } from './scenes/CombatScene';
import { PreloadScene } from './scenes/PreloadScene';

export const GRID_SIZE = 5; // 5x5 grid
export const CELL_SIZE = 80; // pixels per grid cell
export const GAME_WIDTH = GRID_SIZE * CELL_SIZE + 100; // extra padding
export const GAME_HEIGHT = GRID_SIZE * CELL_SIZE + 100;

export const COLORS = {
  PLAYER: 0x10b981,      // green
  ENEMY: 0xef4444,       // red
  NEUTRAL: 0x6b7280,     // gray
  GRID_LINE: 0x334155,   // slate
  BACKGROUND: 0x1e293b,  // dark slate
  HIGHLIGHT: 0x3b82f6,   // blue
  ATTACK: 0xff0000,      // bright red
  HEAL: 0x00ff00,        // bright green
};

export const gameConfig: Phaser.Types.Core.GameConfig = {
  type: Phaser.AUTO,
  width: GAME_WIDTH,
  height: GAME_HEIGHT,
  parent: 'phaser-game',
  backgroundColor: '#1e293b',
  scene: [PreloadScene, CombatScene],
  physics: {
    default: 'arcade',
    arcade: {
      debug: false,
    },
  },
  scale: {
    mode: Phaser.Scale.FIT,
    autoCenter: Phaser.Scale.CENTER_BOTH,
  },
  render: {
    pixelArt: false,
    antialias: true,
    roundPixels: true,
  },
};
