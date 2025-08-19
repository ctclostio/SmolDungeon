import React, { useState } from 'react';
import { Box, Text, useInput } from 'ink';
import type { Character, Action, Id } from '@smol-dungeon/schema';

interface ActionMenuProps {
  character: Character;
  targets: Character[];
  onAction: (action: Action) => void;
}

type MenuState = 'main' | 'attack' | 'ability' | 'item';

export function ActionMenu({ character, targets, onAction }: ActionMenuProps): JSX.Element {
  const [menuState, setMenuState] = useState<MenuState>('main');
  const [selectedIndex, setSelectedIndex] = useState(0);

  useInput((input, key) => {
    if (key.upArrow) {
      setSelectedIndex(prev => Math.max(0, prev - 1));
    } else if (key.downArrow) {
      const maxIndex = getMaxIndex();
      setSelectedIndex(prev => Math.min(maxIndex, prev + 1));
    } else if (key.return) {
      handleSelect();
    } else if (key.escape) {
      setMenuState('main');
      setSelectedIndex(0);
    }
  });

  const getMaxIndex = (): number => {
    switch (menuState) {
      case 'main':
        return 4; // Attack, Defend, Ability, Item, Flee
      case 'attack':
        return Math.max(0, character.weapons.length - 1);
      case 'ability':
        return Math.max(0, character.abilities.length - 1);
      case 'item':
        return Math.max(0, character.items.length - 1);
      default:
        return 0;
    }
  };

  const handleSelect = (): void => {
    switch (menuState) {
      case 'main':
        handleMainMenuSelect();
        break;
      case 'attack':
        handleAttackSelect();
        break;
      case 'ability':
        handleAbilitySelect();
        break;
      case 'item':
        handleItemSelect();
        break;
    }
  };

  const handleMainMenuSelect = (): void => {
    switch (selectedIndex) {
      case 0: // Attack
        if (character.weapons.length > 0) {
          setMenuState('attack');
          setSelectedIndex(0);
        }
        break;
      case 1: // Defend
        onAction({ kind: 'Defend', actor: character.id });
        break;
      case 2: // Ability
        if (character.abilities.length > 0) {
          setMenuState('ability');
          setSelectedIndex(0);
        }
        break;
      case 3: // Item
        if (character.items.length > 0) {
          setMenuState('item');
          setSelectedIndex(0);
        }
        break;
      case 4: // Flee
        onAction({ kind: 'Flee', actor: character.id });
        break;
    }
  };

  const handleAttackSelect = (): void => {
    const weapon = character.weapons[selectedIndex];
    if (weapon && targets.length > 0) {
      // For simplicity, attack the first available target
      const target = targets[0];
      onAction({
        kind: 'Attack',
        attacker: character.id,
        target: target.id,
        weapon: weapon.id,
      });
    }
  };

  const handleAbilitySelect = (): void => {
    const ability = character.abilities[selectedIndex];
    if (ability) {
      const target = targets.length > 0 ? targets[0] : undefined;
      onAction({
        kind: 'Ability',
        actor: character.id,
        ability: ability.id,
        target: target?.id,
      });
    }
  };

  const handleItemSelect = (): void => {
    const item = character.items[selectedIndex];
    if (item) {
      onAction({
        kind: 'UseItem',
        actor: character.id,
        item: item.id,
      });
    }
  };

  const renderMainMenu = (): JSX.Element => (
    <Box flexDirection="column">
      <MenuItem text="Attack" selected={selectedIndex === 0} />
      <MenuItem text="Defend" selected={selectedIndex === 1} />
      <MenuItem text="Ability" selected={selectedIndex === 2} />
      <MenuItem text="Use Item" selected={selectedIndex === 3} />
      <MenuItem text="Flee" selected={selectedIndex === 4} />
    </Box>
  );

  const renderAttackMenu = (): JSX.Element => (
    <Box flexDirection="column">
      <Text color="yellow">Choose weapon:</Text>
      {character.weapons.map((weapon, index) => (
        <MenuItem 
          key={weapon.id}
          text={`${weapon.name} (${weapon.damage} dmg)`}
          selected={selectedIndex === index}
        />
      ))}
    </Box>
  );

  const renderAbilityMenu = (): JSX.Element => (
    <Box flexDirection="column">
      <Text color="yellow">Choose ability:</Text>
      {character.abilities.map((ability, index) => {
        const cooldown = character.abilityCooldowns[ability.id] || 0;
        const available = cooldown === 0;
        return (
          <MenuItem 
            key={ability.id}
            text={`${ability.name}${available ? '' : ` (${cooldown} rounds)`}`}
            selected={selectedIndex === index}
            disabled={!available}
          />
        );
      })}
    </Box>
  );

  const renderItemMenu = (): JSX.Element => (
    <Box flexDirection="column">
      <Text color="yellow">Choose item:</Text>
      {character.items.map((item, index) => (
        <MenuItem 
          key={item.id}
          text={item.name}
          selected={selectedIndex === index}
        />
      ))}
    </Box>
  );

  return (
    <Box flexDirection="column" padding={1}>
      <Text bold color="cyan">{character.name}'s Turn</Text>
      <Text color="gray">Use arrow keys to navigate, Enter to select, Escape to go back</Text>
      
      {menuState === 'main' && renderMainMenu()}
      {menuState === 'attack' && renderAttackMenu()}
      {menuState === 'ability' && renderAbilityMenu()}
      {menuState === 'item' && renderItemMenu()}
    </Box>
  );
}

interface MenuItemProps {
  text: string;
  selected: boolean;
  disabled?: boolean;
}

function MenuItem({ text, selected, disabled = false }: MenuItemProps): JSX.Element {
  const color = disabled ? 'gray' : selected ? 'yellow' : 'white';
  return (
    <Text color={color}>
      {selected ? '► ' : '  '}{text}
    </Text>
  );
}