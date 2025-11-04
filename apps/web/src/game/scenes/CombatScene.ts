import Phaser from 'phaser';
import { Character, GameState, Position } from '../../types/game';
import { GRID_SIZE, CELL_SIZE, COLORS } from '../config';

export class CombatScene extends Phaser.Scene {
  private gridGraphics!: Phaser.GameObjects.Graphics;
  private characterSprites: Map<string, Phaser.GameObjects.Container> = new Map();
  private damageTexts: Phaser.GameObjects.Text[] = [];
  private currentState: GameState | null = null;

  constructor() {
    super({ key: 'CombatScene' });
  }

  create() {
    // Create grid background
    this.createGrid();

    // Initialize event listeners
    this.setupEventListeners();
  }

  private createGrid() {
    const offsetX = 50;
    const offsetY = 50;

    this.gridGraphics = this.add.graphics();
    this.gridGraphics.lineStyle(2, COLORS.GRID_LINE, 0.5);

    // Draw vertical lines
    for (let x = 0; x <= GRID_SIZE; x++) {
      this.gridGraphics.lineBetween(
        offsetX + x * CELL_SIZE,
        offsetY,
        offsetX + x * CELL_SIZE,
        offsetY + GRID_SIZE * CELL_SIZE
      );
    }

    // Draw horizontal lines
    for (let y = 0; y <= GRID_SIZE; y++) {
      this.gridGraphics.lineBetween(
        offsetX,
        offsetY + y * CELL_SIZE,
        offsetX + GRID_SIZE * CELL_SIZE,
        offsetY + y * CELL_SIZE
      );
    }

    // Add coordinate labels
    this.gridGraphics.setDepth(-1);
  }

  private setupEventListeners() {
    // Listen for custom events from React
    this.game.events.on('stateUpdate', (state: GameState) => {
      this.updateGameState(state);
    });

    this.game.events.on('characterDamaged', (data: { characterId: string; amount: number }) => {
      this.showDamageNumber(data.characterId, data.amount, false);
    });

    this.game.events.on('characterHealed', (data: { characterId: string; amount: number }) => {
      this.showDamageNumber(data.characterId, data.amount, true);
    });

    this.game.events.on('characterMissed', (characterId: string) => {
      this.showMissText(characterId);
    });

    this.game.events.on('attackAnimation', (data: { attackerId: string; targetId: string }) => {
      this.playAttackAnimation(data.attackerId, data.targetId);
    });
  }

  updateGameState(state: GameState) {
    this.currentState = state;

    // Remove sprites for dead characters
    const deadCharacterIds = new Set<string>();
    this.characterSprites.forEach((sprite, id) => {
      const character = state.characters.find((c) => c.id === id);
      if (!character || character.stats.hp <= 0) {
        sprite.destroy();
        this.characterSprites.delete(id);
        deadCharacterIds.add(id);
      }
    });

    // Update or create sprites for all living characters
    state.characters.forEach((character) => {
      if (character.stats.hp > 0) {
        this.updateCharacterSprite(character, state);
      }
    });
  }

  private updateCharacterSprite(character: Character, state: GameState) {
    const offsetX = 50;
    const offsetY = 50;

    // Convert game coordinates (centered on 0,0) to screen coordinates
    const screenX = offsetX + (character.position.x + 2) * CELL_SIZE + CELL_SIZE / 2;
    const screenY = offsetY + (character.position.y + 2) * CELL_SIZE + CELL_SIZE / 2;

    let container = this.characterSprites.get(character.id);

    if (!container) {
      // Create new character sprite
      container = this.createCharacterContainer(character, screenX, screenY);
      this.characterSprites.set(character.id, container);
    } else {
      // Update position with smooth animation
      this.tweens.add({
        targets: container,
        x: screenX,
        y: screenY,
        duration: 300,
        ease: 'Power2',
      });
    }

    // Update active indicator
    const isActiveCharacter = state.turnOrder[state.currentTurn] === character.id;
    this.updateActiveIndicator(container, isActiveCharacter);
  }

  private createCharacterContainer(
    character: Character,
    x: number,
    y: number
  ): Phaser.GameObjects.Container {
    const container = this.add.container(x, y);

    // Character circle
    const color = character.isPlayer ? COLORS.PLAYER : COLORS.ENEMY;
    const circle = this.add.circle(0, 0, 25, color);
    circle.setStrokeStyle(3, 0xffffff);

    // Character name
    const nameText = this.add.text(0, -40, character.name, {
      fontSize: '12px',
      color: '#ffffff',
      fontStyle: 'bold',
      backgroundColor: '#00000088',
      padding: { x: 4, y: 2 },
    });
    nameText.setOrigin(0.5);

    // HP text
    const hpText = this.add.text(0, 35, `${character.stats.hp}/${character.stats.maxHp}`, {
      fontSize: '10px',
      color: '#ffffff',
      backgroundColor: '#00000088',
      padding: { x: 3, y: 1 },
    });
    hpText.setOrigin(0.5);

    // Active indicator (initially invisible)
    const activeIndicator = this.add.circle(0, 0, 35, COLORS.HIGHLIGHT, 0);
    activeIndicator.setStrokeStyle(3, COLORS.HIGHLIGHT, 0.8);
    activeIndicator.setName('activeIndicator');

    container.add([activeIndicator, circle, nameText, hpText]);
    container.setDepth(1);
    container.setData('character', character);

    // Add pulse animation
    this.tweens.add({
      targets: circle,
      scaleX: 1.05,
      scaleY: 1.05,
      duration: 1000,
      yoyo: true,
      repeat: -1,
      ease: 'Sine.easeInOut',
    });

    return container;
  }

  private updateActiveIndicator(container: Phaser.GameObjects.Container, isActive: boolean) {
    const indicator = container.getByName('activeIndicator') as Phaser.GameObjects.Arc;

    if (isActive) {
      // Show pulsing indicator
      this.tweens.add({
        targets: indicator,
        alpha: 0.8,
        scaleX: 1.2,
        scaleY: 1.2,
        duration: 500,
        yoyo: true,
        repeat: -1,
        ease: 'Sine.easeInOut',
      });
    } else {
      // Hide indicator
      this.tweens.killTweensOf(indicator);
      indicator.setAlpha(0);
      indicator.setScale(1);
    }
  }

  private showDamageNumber(characterId: string, amount: number, isHeal: boolean) {
    const container = this.characterSprites.get(characterId);
    if (!container) return;

    const color = isHeal ? '#22c55e' : '#ef4444';
    const text = this.add.text(container.x, container.y - 20, `${isHeal ? '+' : '-'}${amount}`, {
      fontSize: '24px',
      color: color,
      fontStyle: 'bold',
      stroke: '#000000',
      strokeThickness: 4,
    });
    text.setOrigin(0.5);
    text.setDepth(10);

    this.tweens.add({
      targets: text,
      y: text.y - 80,
      alpha: 0,
      scale: 1.5,
      duration: 1500,
      ease: 'Power2',
      onComplete: () => {
        text.destroy();
      },
    });
  }

  private showMissText(characterId: string) {
    const container = this.characterSprites.get(characterId);
    if (!container) return;

    const text = this.add.text(container.x, container.y - 20, 'MISS!', {
      fontSize: '20px',
      color: '#9ca3af',
      fontStyle: 'bold',
      stroke: '#000000',
      strokeThickness: 3,
    });
    text.setOrigin(0.5);
    text.setDepth(10);

    this.tweens.add({
      targets: text,
      y: text.y - 60,
      alpha: 0,
      duration: 1200,
      ease: 'Power2',
      onComplete: () => {
        text.destroy();
      },
    });
  }

  private playAttackAnimation(attackerId: string, targetId: string) {
    const attackerContainer = this.characterSprites.get(attackerId);
    const targetContainer = this.characterSprites.get(targetId);

    if (!attackerContainer || !targetContainer) return;

    // Create attack particle/slash effect
    const graphics = this.add.graphics();
    graphics.lineStyle(4, COLORS.ATTACK, 1);
    graphics.lineBetween(
      attackerContainer.x,
      attackerContainer.y,
      targetContainer.x,
      targetContainer.y
    );
    graphics.setDepth(5);

    // Fade out the line
    this.tweens.add({
      targets: graphics,
      alpha: 0,
      duration: 300,
      onComplete: () => {
        graphics.destroy();
      },
    });

    // Shake the target
    const originalX = targetContainer.x;
    this.tweens.add({
      targets: targetContainer,
      x: originalX + 5,
      duration: 50,
      yoyo: true,
      repeat: 3,
      onComplete: () => {
        targetContainer.x = originalX;
      },
    });

    // Flash the target
    const circle = targetContainer.getAll()[1] as Phaser.GameObjects.Arc;
    this.tweens.add({
      targets: circle,
      alpha: 0.3,
      duration: 100,
      yoyo: true,
      repeat: 2,
    });
  }

  shutdown() {
    // Cleanup
    this.game.events.off('stateUpdate');
    this.game.events.off('characterDamaged');
    this.game.events.off('characterHealed');
    this.game.events.off('characterMissed');
    this.game.events.off('attackAnimation');
  }
}
