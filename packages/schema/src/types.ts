import { z } from 'zod';

export const IdSchema = z.string().uuid();
export type Id = z.infer<typeof IdSchema>;

export const StatSchema = z.object({
  hp: z.number().min(0),
  maxHp: z.number().min(1),
  attack: z.number().min(0),
  defense: z.number().min(0),
  speed: z.number().min(0),
});
export type Stat = z.infer<typeof StatSchema>;

export const WeaponSchema = z.object({
  id: IdSchema,
  name: z.string(),
  damage: z.number().min(0),
  accuracy: z.number().min(0).max(100),
});
export type Weapon = z.infer<typeof WeaponSchema>;

export const AbilitySchema = z.object({
  id: IdSchema,
  name: z.string(),
  cooldown: z.number().min(0),
  effect: z.enum(['damage', 'heal', 'buff', 'debuff']),
  power: z.number(),
});
export type Ability = z.infer<typeof AbilitySchema>;

export const ItemSchema = z.object({
  id: IdSchema,
  name: z.string(),
  type: z.enum(['consumable', 'equipment']),
  effect: z.string(),
});
export type Item = z.infer<typeof ItemSchema>;

export const CharacterSchema = z.object({
  id: IdSchema,
  name: z.string(),
  stats: StatSchema,
  weapons: z.array(WeaponSchema),
  abilities: z.array(AbilitySchema),
  items: z.array(ItemSchema),
  abilityCooldowns: z.record(z.string(), z.number()),
  isPlayer: z.boolean(),
});
export type Character = z.infer<typeof CharacterSchema>;

export const ActionSchema = z.discriminatedUnion('kind', [
  z.object({
    kind: z.literal('Attack'),
    attacker: IdSchema,
    target: IdSchema,
    weapon: IdSchema,
  }),
  z.object({
    kind: z.literal('Defend'),
    actor: IdSchema,
  }),
  z.object({
    kind: z.literal('Ability'),
    actor: IdSchema,
    ability: IdSchema,
    target: IdSchema.optional(),
  }),
  z.object({
    kind: z.literal('UseItem'),
    actor: IdSchema,
    item: IdSchema,
  }),
  z.object({
    kind: z.literal('Flee'),
    actor: IdSchema,
  }),
]);
export type Action = z.infer<typeof ActionSchema>;

export const EventSchema = z.discriminatedUnion('type', [
  z.object({
    type: z.literal('damage'),
    target: IdSchema,
    amount: z.number().min(0),
    source: IdSchema.optional(),
  }),
  z.object({
    type: z.literal('heal'),
    target: IdSchema,
    amount: z.number().min(0),
  }),
  z.object({
    type: z.literal('death'),
    target: IdSchema,
  }),
  z.object({
    type: z.literal('flee'),
    actor: IdSchema,
  }),
  z.object({
    type: z.literal('ability_used'),
    actor: IdSchema,
    ability: IdSchema,
    target: IdSchema.optional(),
  }),
  z.object({
    type: z.literal('item_used'),
    actor: IdSchema,
    item: IdSchema,
  }),
]);
export type Event = z.infer<typeof EventSchema>;

export const StateSchema = z.object({
  round: z.number().min(0),
  characters: z.array(CharacterSchema),
  turnOrder: z.array(IdSchema),
  currentTurn: z.number().min(0),
  isComplete: z.boolean(),
  winner: z.enum(['player', 'enemy', 'draw']).optional(),
});
export type State = z.infer<typeof StateSchema>;

export const ResolutionSchema = z.object({
  events: z.array(EventSchema),
  state: StateSchema,
  logs: z.array(z.string()),
});
export type Resolution = z.infer<typeof ResolutionSchema>;

export const RollCheckSchema = z.object({
  actor: IdSchema,
  type: z.enum(['attack', 'defense', 'skill', 'save']),
  dc: z.number().min(1),
});
export type RollCheck = z.infer<typeof RollCheckSchema>;

export const RollResultSchema = z.object({
  roll: z.number().min(1).max(20),
  modifier: z.number(),
  total: z.number(),
  success: z.boolean(),
});
export type RollResult = z.infer<typeof RollResultSchema>;