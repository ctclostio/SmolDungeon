import React, { useState, useEffect } from 'react';
import { Box, Text, useApp } from 'ink';
import { randomUUID } from 'crypto';
import { applyAction, createInitialState, getCurrentCharacter, checkCombatEnd } from '@smol-dungeon/core';
import { EventStore } from '@smol-dungeon/persistence';
import { getDefaultScenario, createCharacterFromTemplate, LLMClient } from '@smol-dungeon/adapters';
import type { State, Action, Character } from '@smol-dungeon/schema';
import { GameState } from './components/GameState.js';
import { ActionMenu } from './components/ActionMenu.js';
import { CombatLog } from './components/CombatLog.js';

interface GameProps {
  dbPath?: string;
  seed?: number;
}

export function Game({ dbPath = ':memory:', seed = Date.now() }: GameProps): JSX.Element {
  const { exit } = useApp();
  const [state, setState] = useState<State | null>(null);
  const [logs, setLogs] = useState<string[]>([]);
  const [eventStore] = useState(() => new EventStore(dbPath));
  const [sessionId] = useState(() => randomUUID());
  const [currentSeed, setCurrentSeed] = useState(seed);
  const [isPlayerTurn, setIsPlayerTurn] = useState(false);

  useEffect(() => {
    initializeGame();
    return () => eventStore.close();
  }, []);

  useEffect(() => {
    if (state && !state.isComplete) {
      const currentChar = getCurrentCharacter(state);
      if (currentChar) {
        setIsPlayerTurn(currentChar.isPlayer);
        
        if (!currentChar.isPlayer) {
          setTimeout(() => handleEnemyTurn(currentChar), 1000);
        }
      }
    }
  }, [state?.currentTurn, state?.round]);

  const initializeGame = async (): Promise<void> => {
    try {
      const scenario = getDefaultScenario();
      const players = scenario.players.map(template => 
        createCharacterFromTemplate(template, true)
      );
      const enemies = scenario.enemies.map(template => 
        createCharacterFromTemplate(template, false)
      );

      const initialState = createInitialState(players, enemies, currentSeed);
      
      await eventStore.createSession(sessionId, scenario.name);
      await eventStore.saveSnapshot(sessionId, 0, initialState);
      
      setState(initialState);
      setLogs([`${scenario.name} begins!`, scenario.context]);
    } catch (error) {
      console.error('Failed to initialize game:', error);
      setLogs(['Failed to initialize game']);
    }
  };

  const handlePlayerAction = async (action: Action): Promise<void> => {
    if (!state || state.isComplete) return;

    try {
      const resolution = applyAction(state, action, currentSeed + state.round);
      
      await eventStore.appendEvents(sessionId, state.round, resolution.events);
      
      if (resolution.state.round !== state.round) {
        await eventStore.saveSnapshot(sessionId, resolution.state.round, resolution.state);
      }
      
      setState(resolution.state);
      setLogs(prev => [...prev, ...resolution.logs]);
      setCurrentSeed(prev => prev + 1);
      
      if (checkCombatEnd(resolution.state)) {
        await eventStore.updateSessionStatus(sessionId, 'completed');
        setTimeout(() => exit(), 3000);
      }
    } catch (error) {
      console.error('Failed to process action:', error);
      setLogs(prev => [...prev, 'Failed to process action']);
    }
  };

  const handleEnemyTurn = async (enemy: Character): Promise<void> => {
    if (!state || state.isComplete) return;

    try {
      const aliveTargets = state.characters.filter(c => c.isPlayer && c.stats.hp > 0);
      if (aliveTargets.length === 0) return;

      let action: Action;
      
      if (enemy.stats.hp < enemy.stats.maxHp / 3 && enemy.items.length > 0) {
        action = {
          kind: 'UseItem',
          actor: enemy.id,
          item: enemy.items[0].id,
        };
      } else if (enemy.abilities.length > 0) {
        const availableAbilities = enemy.abilities.filter(ability => 
          (enemy.abilityCooldowns[ability.id] || 0) === 0
        );
        
        if (availableAbilities.length > 0) {
          action = {
            kind: 'Ability',
            actor: enemy.id,
            ability: availableAbilities[0].id,
            target: aliveTargets[0].id,
          };
        } else {
          action = {
            kind: 'Attack',
            attacker: enemy.id,
            target: aliveTargets[0].id,
            weapon: enemy.weapons[0].id,
          };
        }
      } else {
        action = {
          kind: 'Attack',
          attacker: enemy.id,
          target: aliveTargets[0].id,
          weapon: enemy.weapons[0].id,
        };
      }

      await handlePlayerAction(action);
    } catch (error) {
      console.error('Enemy turn failed:', error);
    }
  };

  if (!state) {
    return (
      <Box padding={1}>
        <Text>Loading game...</Text>
      </Box>
    );
  }

  const currentChar = getCurrentCharacter(state);
  const targets = state.characters.filter(c => 
    currentChar?.isPlayer ? !c.isPlayer && c.stats.hp > 0 : c.isPlayer && c.stats.hp > 0
  );

  return (
    <Box flexDirection="column" padding={1}>
      <Text bold color="blue">Smol Dungeon - Combat Encounter</Text>
      
      <GameState state={state} />
      
      <Box flexDirection="row" gap={2}>
        <Box flexGrow={1}>
          <CombatLog logs={logs} maxLines={8} />
        </Box>
        
        {isPlayerTurn && currentChar && (
          <Box width={30}>
            <ActionMenu 
              character={currentChar}
              targets={targets}
              onAction={handlePlayerAction}
            />
          </Box>
        )}
      </Box>
      
      {state.isComplete && (
        <Box marginTop={1}>
          <Text color="gray">Game will exit in 3 seconds...</Text>
        </Box>
      )}
    </Box>
  );
}