import type { State, Character } from '@smol-dungeon/schema';

interface CombatMapProps {
  state: State;
}

export function CombatMap({ state }: CombatMapProps): JSX.Element {
  const gridSize = 5;
  const offset = Math.floor(gridSize / 2);

  // Create a map of positions to characters
  const positionMap = new Map<string, Character>();
  state.characters.forEach(char => {
    const key = `${char.position.x},${char.position.y}`;
    positionMap.set(key, char);
  });

  const currentChar = state.characters[state.currentTurn];

  const renderCell = (x: number, y: number): JSX.Element => {
    const key = `${x},${y}`;
    const character = positionMap.get(key);

    let cellContent = '';
    let cellStyle: React.CSSProperties = {
      width: '50px',
      height: '50px',
      border: '1px solid #ddd',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: '11px',
      fontWeight: 'bold',
      position: 'relative',
      borderRadius: '4px',
      boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
    };

    if (character) {
      const isCurrent = currentChar?.id === character.id;
      const isDead = character.stats.hp === 0;
      const isLowHealth = character.stats.hp < character.stats.maxHp / 3;

      cellStyle.backgroundColor = character.isPlayer ? '#4CAF50' : '#F44336';
      cellStyle.color = 'white';
      cellStyle.border = isCurrent ? '3px solid #2196F3' : '2px solid #333';
      cellStyle.boxShadow = isCurrent ? '0 0 10px rgba(33, 150, 243, 0.5)' : '0 2px 4px rgba(0,0,0,0.1)';

      if (isDead) {
        cellStyle.backgroundColor = '#9E9E9E';
        cellStyle.opacity = 0.6;
      } else if (isLowHealth) {
        cellStyle.backgroundColor = '#FF9800';
        cellStyle.color = 'black';
      }

      cellContent = character.name.split(' ')[0]; // First word of name

      // Add HP indicator
      const hpPercent = (character.stats.hp / character.stats.maxHp) * 100;
      const hpBarStyle: React.CSSProperties = {
        position: 'absolute',
        bottom: '3px',
        left: '3px',
        right: '3px',
        height: '5px',
        backgroundColor: 'rgba(0,0,0,0.3)',
        borderRadius: '2px',
      };

      const hpFillStyle: React.CSSProperties = {
        height: '100%',
        width: `${hpPercent}%`,
        backgroundColor: hpPercent > 60 ? '#4CAF50' : hpPercent > 30 ? '#FF9800' : '#F44336',
        borderRadius: '2px',
        transition: 'width 0.3s ease',
      };

      return (
        <div
          key={key}
          style={cellStyle}
          title={`${character.name} (${character.position.x}, ${character.position.y})\nHP: ${character.stats.hp}/${character.stats.maxHp}\n${isCurrent ? 'Current Turn' : ''}`}
        >
          {cellContent}
          <div style={hpBarStyle}>
            <div style={hpFillStyle}></div>
          </div>
        </div>
      );
    } else {
      cellStyle.backgroundColor = '#E8F5E8';
      cellStyle.border = '1px solid #C8E6C9';
      return <div key={key} style={cellStyle}></div>;
    }
  };

  const gridCells = [];
  for (let y = -offset; y <= offset; y++) {
    for (let x = -offset; x <= offset; x++) {
      gridCells.push(renderCell(x, y));
    }
  }

  return (
    <div style={{ marginBottom: '20px' }}>
      <h3 style={{ color: '#007bff', margin: '0 0 10px 0', textAlign: 'center' }}>Combat Map</h3>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: `repeat(${gridSize}, 1fr)`,
          gap: '2px',
          justifyContent: 'center',
          maxWidth: '320px',
          margin: '0 auto',
        }}
      >
        {gridCells}
      </div>
      <div style={{ textAlign: 'center', marginTop: '10px', fontSize: '12px', color: '#666' }}>
        <div><strong>Legend:</strong></div>
        <div>🟢 Player | 🔴 Enemy | 🔵 Current Turn | 🟡 Low HP | ⚫ Dead</div>
        <div>Hover for details • HP bar shows health percentage</div>
      </div>
    </div>
  );
}