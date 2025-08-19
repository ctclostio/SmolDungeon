import { sqliteTable, text, integer, blob } from 'drizzle-orm/sqlite-core';

export const events = sqliteTable('events', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  sessionId: text('session_id').notNull(),
  round: integer('round').notNull(),
  eventData: text('event_data', { mode: 'json' }).notNull(),
  timestamp: integer('timestamp', { mode: 'timestamp' }).notNull().default(new Date()),
});

export const snapshots = sqliteTable('snapshots', {
  id: integer('id').primaryKey({ autoIncrement: true }),
  sessionId: text('session_id').notNull(),
  round: integer('round').notNull(),
  stateData: text('state_data', { mode: 'json' }).notNull(),
  timestamp: integer('timestamp', { mode: 'timestamp' }).notNull().default(new Date()),
});

export const sessions = sqliteTable('sessions', {
  id: text('id').primaryKey(),
  name: text('name').notNull(),
  status: text('status').notNull().default('active'),
  createdAt: integer('created_at', { mode: 'timestamp' }).notNull().default(new Date()),
  updatedAt: integer('updated_at', { mode: 'timestamp' }).notNull().default(new Date()),
});