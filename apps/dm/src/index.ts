import Fastify from 'fastify';
import { randomUUID } from 'crypto';
import { applyAction, getStateSummary, SeededRNG } from '@smol-dungeon/core';
import { EventStore } from '@smol-dungeon/persistence';
import { ActionSchema, RollCheckSchema, type Action, type State, type RollCheck, type RollResult } from '@smol-dungeon/schema';

const fastify = Fastify({ logger: true });

const dbPath = process.env.DB_PATH || './dm-server.db';
const port = parseInt(process.env.PORT || '3000', 10);
const eventStore = new EventStore(dbPath);

interface DMState {
  [sessionId: string]: State;
}

const currentStates: DMState = {};

fastify.post<{
  Body: { state: State };
  Reply: { summary: string };
}>('/tools/get_state_summary', async (request, reply) => {
  try {
    const { state } = request.body;
    
    if (!state) {
      return reply.code(400).send({ error: 'State is required' });
    }

    const summary = getStateSummary(state);
    
    return { summary };
  } catch (error) {
    fastify.log.error('Error in get_state_summary:', error);
    return reply.code(500).send({ error: 'Internal server error' });
  }
});

fastify.post<{
  Body: RollCheck;
  Reply: RollResult;
}>('/tools/roll_check', async (request, reply) => {
  try {
    const rollCheck = RollCheckSchema.parse(request.body);
    
    const seed = Date.now();
    const rng = new SeededRNG(seed);
    const roll = rng.rollD20();
    
    let modifier = 0;
    const sessionId = request.headers['session-id'] as string;
    
    if (sessionId && currentStates[sessionId]) {
      const state = currentStates[sessionId];
      const character = state.characters.find(c => c.id === rollCheck.actor);
      
      if (character) {
        switch (rollCheck.type) {
          case 'attack':
            modifier = character.stats.attack;
            break;
          case 'defense':
            modifier = character.stats.defense;
            break;
          case 'skill':
          case 'save':
            modifier = Math.floor(character.stats.speed / 2);
            break;
        }
      }
    }
    
    const total = roll + modifier;
    const success = total >= rollCheck.dc;
    
    return {
      roll,
      modifier,
      total,
      success,
    };
  } catch (error) {
    fastify.log.error('Error in roll_check:', error);
    return reply.code(500).send({ error: 'Internal server error' });
  }
});

fastify.post<{
  Body: { state: State; action: Action; seed: number };
  Reply: { events: any[]; state: State; logs: string[] };
}>('/tools/apply_action', async (request, reply) => {
  try {
    const { state, action, seed } = request.body;
    
    if (!state || !action || seed === undefined) {
      return reply.code(400).send({ 
        error: 'State, action, and seed are required' 
      });
    }

    const validatedAction = ActionSchema.parse(action);
    const resolution = applyAction(state, validatedAction, seed);
    
    const sessionId = request.headers['session-id'] as string || randomUUID();
    currentStates[sessionId] = resolution.state;
    
    await eventStore.appendEvents(sessionId, resolution.state.round, resolution.events);
    
    if (resolution.state.round > state.round) {
      await eventStore.saveSnapshot(sessionId, resolution.state.round, resolution.state);
    }
    
    return resolution;
  } catch (error) {
    fastify.log.error('Error in apply_action:', error);
    return reply.code(500).send({ error: 'Internal server error' });
  }
});

fastify.get('/health', async () => {
  return { status: 'ok', timestamp: new Date().toISOString() };
});

fastify.get('/sessions', async () => {
  return { 
    activeSessions: Object.keys(currentStates),
    totalSessions: Object.keys(currentStates).length,
  };
});

fastify.post<{
  Body: { sessionId: string; state: State };
}>('/sessions', async (request, reply) => {
  try {
    const { sessionId, state } = request.body;
    
    if (!sessionId || !state) {
      return reply.code(400).send({ 
        error: 'Session ID and state are required' 
      });
    }
    
    currentStates[sessionId] = state;
    await eventStore.createSession(sessionId, `Session ${sessionId}`);
    await eventStore.saveSnapshot(sessionId, state.round, state);
    
    return { success: true, sessionId };
  } catch (error) {
    fastify.log.error('Error creating session:', error);
    return reply.code(500).send({ error: 'Internal server error' });
  }
});

fastify.get<{
  Params: { sessionId: string };
}>('/sessions/:sessionId', async (request, reply) => {
  try {
    const { sessionId } = request.params;
    
    const state = currentStates[sessionId];
    if (!state) {
      return reply.code(404).send({ error: 'Session not found' });
    }
    
    return { sessionId, state };
  } catch (error) {
    fastify.log.error('Error getting session:', error);
    return reply.code(500).send({ error: 'Internal server error' });
  }
});

const start = async (): Promise<void> => {
  try {
    await fastify.listen({ port, host: '0.0.0.0' });
    console.log(`DM Server running on http://localhost:${port}`);
    console.log(`Database: ${dbPath}`);
    console.log('\nAvailable endpoints:');
    console.log('POST /tools/get_state_summary');
    console.log('POST /tools/roll_check');
    console.log('POST /tools/apply_action');
    console.log('GET  /health');
    console.log('GET  /sessions');
    console.log('POST /sessions');
    console.log('GET  /sessions/:sessionId');
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
};

process.on('SIGINT', async () => {
  console.log('\nShutting down DM server...');
  eventStore.close();
  await fastify.close();
  process.exit(0);
});

start();