import { useState } from 'react';
import type { Character, Action } from '@smol-dungeon/schema';

interface ActionMenuProps {
  character: Character;
  targets: Character[];
  onAction: (action: Action) => void;
}

type MenuState = 'main' | 'attack' | 'ability' | 'item';

export function ActionMenu({ character, targets, onAction }: ActionMenuProps): JSX.Element {
  const [menuState, setMenuState] = useState<MenuState>('main');

  const handleMainMenuSelect = (action: string): void => {
    switch (action) {
      case 'attack':
        if (character.weapons.length > 0) {
          setMenuState('attack');
        }
        break;
      case 'defend':
        onAction({ kind: 'Defend', actor: character.id });
        break;
      case 'ability':
        if (character.abilities.length > 0) {
          setMenuState('ability');
        }
        break;
      case 'item':
        if (character.items.length > 0) {
          setMenuState('item');
        }
        break;
      case 'flee':
        onAction({ kind: 'Flee', actor: character.id });
        break;
    }
  };

  const handleAttackSelect = (weaponId: string): void => {
    if (targets.length > 0) {
      const target = targets[0]; // For simplicity, attack the first available target
      onAction({
        kind: 'Attack',
        attacker: character.id,
        target: target.id,
        weapon: weaponId,
      });
    }
  };

  const handleAbilitySelect = (abilityId: string): void => {
    const ability = character.abilities.find(a => a.id === abilityId);
    if (ability) {
      const target = targets.length > 0 ? targets[0] : undefined;
      onAction({
        kind: 'Ability',
        actor: character.id,
        ability: abilityId,
        target: target?.id,
      });
    }
  };

  const handleItemSelect = (itemId: string): void => {
    onAction({
      kind: 'UseItem',
      actor: character.id,
      item: itemId,
    });
  };

  const renderMainMenu = (): JSX.Element => (
    <div>
      <h4 style={{ margin: '0 0 10px 0', color: '#007bff' }}>
        {character.name}'s Turn
      </h4>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
        <button
          onClick={() => handleMainMenuSelect('attack')}
          disabled={character.weapons.length === 0}
          style={buttonStyle}
        >
          Attack
        </button>
        <button
          onClick={() => handleMainMenuSelect('defend')}
          style={buttonStyle}
        >
          Defend
        </button>
        <button
          onClick={() => handleMainMenuSelect('ability')}
          disabled={character.abilities.length === 0}
          style={buttonStyle}
        >
          Ability
        </button>
        <button
          onClick={() => handleMainMenuSelect('item')}
          disabled={character.items.length === 0}
          style={buttonStyle}
        >
          Use Item
        </button>
        <button
          onClick={() => handleMainMenuSelect('flee')}
          style={buttonStyle}
        >
          Flee
        </button>
      </div>
    </div>
  );

  const renderAttackMenu = (): JSX.Element => (
    <div>
      <h4 style={{ margin: '0 0 10px 0', color: '#ffc107' }}>Choose weapon:</h4>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
        {character.weapons.map((weapon) => (
          <button
            key={weapon.id}
            onClick={() => handleAttackSelect(weapon.id)}
            style={buttonStyle}
          >
            {weapon.name} ({weapon.damage} dmg)
          </button>
        ))}
      </div>
      <button
        onClick={() => setMenuState('main')}
        style={{ ...buttonStyle, marginTop: '10px', backgroundColor: '#6c757d' }}
      >
        Back
      </button>
    </div>
  );

  const renderAbilityMenu = (): JSX.Element => (
    <div>
      <h4 style={{ margin: '0 0 10px 0', color: '#ffc107' }}>Choose ability:</h4>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
        {character.abilities.map((ability) => {
          const cooldown = character.abilityCooldowns[ability.id] || 0;
          const available = cooldown === 0;
          return (
            <button
              key={ability.id}
              onClick={() => available && handleAbilitySelect(ability.id)}
              disabled={!available}
              style={{
                ...buttonStyle,
                opacity: available ? 1 : 0.5,
                cursor: available ? 'pointer' : 'not-allowed'
              }}
            >
              {ability.name}{available ? '' : ` (${cooldown} rounds)`}
            </button>
          );
        })}
      </div>
      <button
        onClick={() => setMenuState('main')}
        style={{ ...buttonStyle, marginTop: '10px', backgroundColor: '#6c757d' }}
      >
        Back
      </button>
    </div>
  );

  const renderItemMenu = (): JSX.Element => (
    <div>
      <h4 style={{ margin: '0 0 10px 0', color: '#ffc107' }}>Choose item:</h4>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
        {character.items.map((item) => (
          <button
            key={item.id}
            onClick={() => handleItemSelect(item.id)}
            style={buttonStyle}
          >
            {item.name}
          </button>
        ))}
      </div>
      <button
        onClick={() => setMenuState('main')}
        style={{ ...buttonStyle, marginTop: '10px', backgroundColor: '#6c757d' }}
      >
        Back
      </button>
    </div>
  );

  return (
    <div style={{
      border: '1px solid #ccc',
      borderRadius: '4px',
      padding: '15px',
      backgroundColor: '#f8f9fa'
    }}>
      {menuState === 'main' && renderMainMenu()}
      {menuState === 'attack' && renderAttackMenu()}
      {menuState === 'ability' && renderAbilityMenu()}
      {menuState === 'item' && renderItemMenu()}
    </div>
  );
}

const buttonStyle: React.CSSProperties = {
  padding: '8px 12px',
  border: '1px solid #007bff',
  borderRadius: '4px',
  backgroundColor: '#007bff',
  color: 'white',
  cursor: 'pointer',
  fontSize: '14px',
  transition: 'background-color 0.2s',
};