import type { Action, Event, Resolution, State, Id } from '@smol-dungeon/schema';
import { SeededRNG } from './rng.js';

export function applyAction(state: State, action: Action, seed: number): Resolution {
  const rng = new SeededRNG(seed);
  const events: Event[] = [];
  const logs: string[] = [];
  
  const newState = structuredClone(state);
  const character = newState.characters.find(c => c.id === getActorId(action));
  
  if (!character) {
    throw new Error(`Character not found: ${getActorId(action)}`);
  }

  switch (action.kind) {
    case 'Attack':
      return handleAttack(newState, action, rng, events, logs);
    case 'Defend':
      return handleDefend(newState, action, rng, events, logs);
    case 'Ability':
      return handleAbility(newState, action, rng, events, logs);
    case 'UseItem':
      return handleUseItem(newState, action, rng, events, logs);
    case 'Flee':
      return handleFlee(newState, action, rng, events, logs);
  }
}

function getActorId(action: Action): Id {
  switch (action.kind) {
    case 'Attack':
      return action.attacker;
    case 'Defend':
    case 'Ability':
    case 'UseItem':
    case 'Flee':
      return action.actor;
  }
}

function handleAttack(
  state: State,
  action: Extract<Action, { kind: 'Attack' }>,
  rng: SeededRNG,
  events: Event[],
  logs: string[]
): Resolution {
  const attacker = state.characters.find(c => c.id === action.attacker);
  const target = state.characters.find(c => c.id === action.target);
  const weapon = attacker?.weapons.find(w => w.id === action.weapon);

  if (!attacker || !target || !weapon) {
    return { events, state, logs: [...logs, 'Invalid attack action'] };
  }

  const attackRoll = rng.rollD20();
  const hit = attackRoll + attacker.stats.attack >= target.stats.defense + 10;

  if (hit) {
    const baseDamage = weapon.damage + Math.floor(attacker.stats.attack / 2);
    const damageRoll = rng.rollD6();
    const totalDamage = Math.max(1, baseDamage + damageRoll - target.stats.defense);

    target.stats.hp = Math.max(0, target.stats.hp - totalDamage);

    events.push({
      type: 'damage',
      target: target.id,
      amount: totalDamage,
      source: attacker.id,
    });

    logs.push(`${attacker.name} attacks ${target.name} with ${weapon.name} for ${totalDamage} damage!`);

    if (target.stats.hp === 0) {
      events.push({
        type: 'death',
        target: target.id,
      });
      logs.push(`${target.name} has been defeated!`);
    }
  } else {
    logs.push(`${attacker.name} misses ${target.name}!`);
  }

  const updatedState = advanceTurn(state);
  return { events, state: updatedState, logs };
}

function handleDefend(
  state: State,
  action: Extract<Action, { kind: 'Defend' }>,
  _rng: SeededRNG,
  events: Event[],
  logs: string[]
): Resolution {
  const character = state.characters.find(c => c.id === action.actor);
  
  if (!character) {
    return { events, state, logs: [...logs, 'Invalid defend action'] };
  }

  character.stats.defense += 2;
  logs.push(`${character.name} takes a defensive stance!`);

  const updatedState = advanceTurn(state);
  return { events, state: updatedState, logs };
}

function handleAbility(
  state: State,
  action: Extract<Action, { kind: 'Ability' }>,
  rng: SeededRNG,
  events: Event[],
  logs: string[]
): Resolution {
  const character = state.characters.find(c => c.id === action.actor);
  const ability = character?.abilities.find(a => a.id === action.ability);

  if (!character || !ability) {
    return { events, state, logs: [...logs, 'Invalid ability action'] };
  }

  const cooldownKey = ability.id;
  const currentCooldown = character.abilityCooldowns[cooldownKey] || 0;
  
  if (currentCooldown > 0) {
    return { events, state, logs: [...logs, `${ability.name} is on cooldown!`] };
  }

  character.abilityCooldowns[cooldownKey] = ability.cooldown;

  events.push({
    type: 'ability_used',
    actor: character.id,
    ability: ability.id,
    target: action.target,
  });

  switch (ability.effect) {
    case 'damage':
      if (action.target) {
        const target = state.characters.find(c => c.id === action.target);
        if (target) {
          const damage = ability.power + rng.rollD6();
          target.stats.hp = Math.max(0, target.stats.hp - damage);
          
          events.push({
            type: 'damage',
            target: target.id,
            amount: damage,
            source: character.id,
          });
          
          logs.push(`${character.name} uses ${ability.name} on ${target.name} for ${damage} damage!`);
          
          if (target.stats.hp === 0) {
            events.push({
              type: 'death',
              target: target.id,
            });
            logs.push(`${target.name} has been defeated!`);
          }
        }
      }
      break;
    case 'heal':
      const healAmount = ability.power + rng.rollD6();
      character.stats.hp = Math.min(character.stats.maxHp, character.stats.hp + healAmount);
      
      events.push({
        type: 'heal',
        target: character.id,
        amount: healAmount,
      });
      
      logs.push(`${character.name} uses ${ability.name} and heals for ${healAmount} HP!`);
      break;
  }

  const updatedState = advanceTurn(state);
  return { events, state: updatedState, logs };
}

function handleUseItem(
  state: State,
  action: Extract<Action, { kind: 'UseItem' }>,
  rng: SeededRNG,
  events: Event[],
  logs: string[]
): Resolution {
  const character = state.characters.find(c => c.id === action.actor);
  const itemIndex = character?.items.findIndex(i => i.id === action.item);

  if (!character || itemIndex === undefined || itemIndex === -1) {
    return { events, state, logs: [...logs, 'Invalid item action'] };
  }

  const item = character.items[itemIndex];
  character.items.splice(itemIndex, 1);

  events.push({
    type: 'item_used',
    actor: character.id,
    item: item.id,
  });

  if (item.name.includes('Potion')) {
    const healAmount = 20 + rng.rollD6();
    character.stats.hp = Math.min(character.stats.maxHp, character.stats.hp + healAmount);
    
    events.push({
      type: 'heal',
      target: character.id,
      amount: healAmount,
    });
    
    logs.push(`${character.name} uses ${item.name} and heals for ${healAmount} HP!`);
  }

  const updatedState = advanceTurn(state);
  return { events, state: updatedState, logs };
}

function handleFlee(
  state: State,
  action: Extract<Action, { kind: 'Flee' }>,
  rng: SeededRNG,
  events: Event[],
  logs: string[]
): Resolution {
  const character = state.characters.find(c => c.id === action.actor);
  
  if (!character) {
    return { events, state, logs: [...logs, 'Invalid flee action'] };
  }

  const fleeRoll = rng.rollD20();
  const success = fleeRoll + character.stats.speed >= 15;

  if (success) {
    events.push({
      type: 'flee',
      actor: character.id,
    });
    
    logs.push(`${character.name} successfully flees from combat!`);
    
    const updatedState = { ...state, isComplete: true, winner: character.isPlayer ? undefined : 'player' as const };
    return { events, state: updatedState, logs };
  } else {
    logs.push(`${character.name} fails to flee!`);
    const updatedState = advanceTurn(state);
    return { events, state: updatedState, logs };
  }
}

function advanceTurn(state: State): State {
  const updatedState = structuredClone(state);
  
  updatedState.characters.forEach(character => {
    Object.keys(character.abilityCooldowns).forEach(abilityId => {
      if (character.abilityCooldowns[abilityId] > 0) {
        character.abilityCooldowns[abilityId]--;
      }
    });
    
    if (character.stats.defense > character.stats.defense) {
      character.stats.defense = Math.max(0, character.stats.defense - 2);
    }
  });

  const alivePlayers = updatedState.characters.filter(c => c.isPlayer && c.stats.hp > 0);
  const aliveEnemies = updatedState.characters.filter(c => !c.isPlayer && c.stats.hp > 0);
  
  if (alivePlayers.length === 0) {
    updatedState.isComplete = true;
    updatedState.winner = 'enemy';
  } else if (aliveEnemies.length === 0) {
    updatedState.isComplete = true;
    updatedState.winner = 'player';
  }

  if (!updatedState.isComplete) {
    updatedState.currentTurn = (updatedState.currentTurn + 1) % updatedState.turnOrder.length;
    
    if (updatedState.currentTurn === 0) {
      updatedState.round++;
    }
  }

  return updatedState;
}

export function checkCombatEnd(state: State): boolean {
  return state.isComplete || state.round >= 20;
}