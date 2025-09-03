import type { State, Character } from '@smol-dungeon/schema';
import { CombatMap } from './CombatMap';

interface GameStateProps {
  state: State;
}

export function GameState({ state }: GameStateProps): JSX.Element {
  const players = state.characters.filter(c => c.isPlayer);
  const enemies = state.characters.filter(c => !c.isPlayer);
  const currentChar = state.characters[state.currentTurn];

  return (
    <div style={{ marginBottom: '20px' }}>
      <CombatMap state={state} />

      <div style={{ marginBottom: '10px' }}>
        <h2 style={{ color: '#007bff', margin: '0' }}>
          Round {state.round} {state.isComplete ? '(COMPLETE)' : ''}
        </h2>
      </div>

      <div style={{ display: 'flex', gap: '40px' }}>
        <div>
          <h3 style={{ color: '#28a745', margin: '0 0 10px 0' }}>Players:</h3>
          {players.map(player => (
            <CharacterInfo
              key={player.id}
              character={player}
              isCurrent={currentChar?.id === player.id}
            />
          ))}
        </div>

        <div>
          <h3 style={{ color: '#dc3545', margin: '0 0 10px 0' }}>Enemies:</h3>
          {enemies.map(enemy => (
            <CharacterInfo
              key={enemy.id}
              character={enemy}
              isCurrent={currentChar?.id === enemy.id}
            />
          ))}
        </div>
      </div>

      {state.isComplete && state.winner && (
        <div style={{ marginTop: '20px', textAlign: 'center' }}>
          <h3 style={{
            color: state.winner === 'player' ? '#28a745' : '#dc3545',
            margin: '0'
          }}>
            Victory: {state.winner === 'player' ? 'Players Win!' : 'Enemies Win!'}
          </h3>
        </div>
      )}
    </div>
  );
}

interface CharacterInfoProps {
  character: Character;
  isCurrent: boolean;
}

function CharacterInfo({ character, isCurrent }: CharacterInfoProps): JSX.Element {
  const hpColor = character.stats.hp === 0 ? '#dc3545' :
                  character.stats.hp < character.stats.maxHp / 3 ? '#ffc107' : '#000';

  return (
    <div style={{
      marginBottom: '5px',
      padding: '5px',
      backgroundColor: isCurrent ? '#e3f2fd' : 'transparent',
      borderRadius: '4px'
    }}>
      <div style={{ color: isCurrent ? '#007bff' : '#000' }}>
        {isCurrent ? '► ' : '  '}
        <strong>{character.name}:</strong>
        <span style={{ color: hpColor, marginLeft: '5px' }}>
          {character.stats.hp}/{character.stats.maxHp} HP
        </span>
      </div>
    </div>
  );
}