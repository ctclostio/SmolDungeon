#!/usr/bin/env node
import React from 'react';
import { render } from 'ink';
import { Game } from './Game.js';

const dbPath = process.env.DB_PATH || './smol-dungeon.db';
const seed = process.env.SEED ? parseInt(process.env.SEED, 10) : Date.now();

render(React.createElement(Game, { dbPath, seed }));