import React from 'react';
import { Box, Text } from 'ink';
import type { State, Character } from '@smol-dungeon/schema';

interface GameStateProps {
  state: State;
}

export function GameState({ state }: GameStateProps): JSX.Element {
  const players = state.characters.filter(c => c.isPlayer);
  const enemies = state.characters.filter(c => !c.isPlayer);
  const currentChar = state.characters[state.currentTurn];

  return (
    <Box flexDirection="column" padding={1}>
      <Box marginBottom={1}>
        <Text bold color="blue">
          Round {state.round} {state.isComplete ? '(COMPLETE)' : ''}
        </Text>
      </Box>
      
      <Box flexDirection="row" gap={4}>
        <Box flexDirection="column">
          <Text bold color="green">Players:</Text>
          {players.map(player => (
            <CharacterInfo 
              key={player.id} 
              character={player} 
              isCurrent={currentChar?.id === player.id}
            />
          ))}
        </Box>
        
        <Box flexDirection="column">
          <Text bold color="red">Enemies:</Text>
          {enemies.map(enemy => (
            <CharacterInfo 
              key={enemy.id} 
              character={enemy} 
              isCurrent={currentChar?.id === enemy.id}
            />
          ))}
        </Box>
      </Box>
      
      {state.isComplete && state.winner && (
        <Box marginTop={1}>
          <Text bold color={state.winner === 'player' ? 'green' : 'red'}>
            Victory: {state.winner === 'player' ? 'Players Win!' : 'Enemies Win!'}
          </Text>
        </Box>
      )}
    </Box>
  );
}

interface CharacterInfoProps {
  character: Character;
  isCurrent: boolean;
}

function CharacterInfo({ character, isCurrent }: CharacterInfoProps): JSX.Element {
  const hpColor = character.stats.hp === 0 ? 'red' : 
                  character.stats.hp < character.stats.maxHp / 3 ? 'yellow' : 'white';
  
  return (
    <Box>
      <Text color={isCurrent ? 'cyan' : 'white'}>
        {isCurrent ? '► ' : '  '}
        {character.name}: 
        <Text color={hpColor}> {character.stats.hp}/{character.stats.maxHp} HP</Text>
      </Text>
    </Box>
  );
}