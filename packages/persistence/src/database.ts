import Database from 'better-sqlite3';
import { drizzle } from 'drizzle-orm/better-sqlite3';
import { migrate } from 'drizzle-orm/better-sqlite3/migrator';
import { eq, desc } from 'drizzle-orm';
import type { Event, State } from '@smol-dungeon/schema';
import { events, snapshots, sessions } from './schema.js';

export class EventStore {
  private db: Database.Database;
  private drizzle: ReturnType<typeof drizzle>;

  constructor(dbPath: string = ':memory:') {
    this.db = new Database(dbPath);
    this.drizzle = drizzle(this.db);
    
    this.db.exec(`
      CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        session_id TEXT NOT NULL,
        round INTEGER NOT NULL,
        event_data TEXT NOT NULL,
        timestamp INTEGER NOT NULL DEFAULT (unixepoch())
      );
      
      CREATE TABLE IF NOT EXISTS snapshots (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        session_id TEXT NOT NULL,
        round INTEGER NOT NULL,
        state_data TEXT NOT NULL,
        timestamp INTEGER NOT NULL DEFAULT (unixepoch())
      );
      
      CREATE TABLE IF NOT EXISTS sessions (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'active',
        created_at INTEGER NOT NULL DEFAULT (unixepoch()),
        updated_at INTEGER NOT NULL DEFAULT (unixepoch())
      );
      
      CREATE INDEX IF NOT EXISTS idx_events_session_round ON events(session_id, round);
      CREATE INDEX IF NOT EXISTS idx_snapshots_session_round ON snapshots(session_id, round);
    `);
  }

  async createSession(sessionId: string, name: string): Promise<void> {
    await this.drizzle.insert(sessions).values({
      id: sessionId,
      name,
      status: 'active',
    });
  }

  async appendEvents(sessionId: string, round: number, eventList: Event[]): Promise<void> {
    const values = eventList.map(event => ({
      sessionId,
      round,
      eventData: event,
    }));

    if (values.length > 0) {
      await this.drizzle.insert(events).values(values);
    }
  }

  async saveSnapshot(sessionId: string, round: number, state: State): Promise<void> {
    await this.drizzle.insert(snapshots).values({
      sessionId,
      round,
      stateData: state,
    });
  }

  async getEvents(sessionId: string, fromRound: number = 0): Promise<Event[]> {
    const result = await this.drizzle
      .select()
      .from(events)
      .where(eq(events.sessionId, sessionId))
      .orderBy(events.round, events.id);

    return result
      .filter(row => row.round >= fromRound)
      .map(row => row.eventData as Event);
  }

  async getLatestSnapshot(sessionId: string): Promise<State | null> {
    const result = await this.drizzle
      .select()
      .from(snapshots)
      .where(eq(snapshots.sessionId, sessionId))
      .orderBy(desc(snapshots.round))
      .limit(1);

    return result.length > 0 ? result[0].stateData as State : null;
  }

  async getSnapshotAtRound(sessionId: string, round: number): Promise<State | null> {
    const result = await this.drizzle
      .select()
      .from(snapshots)
      .where(eq(snapshots.sessionId, sessionId))
      .orderBy(desc(snapshots.round))
      .limit(1);

    const snapshot = result.find(s => s.round <= round);
    return snapshot ? snapshot.stateData as State : null;
  }

  async updateSessionStatus(sessionId: string, status: string): Promise<void> {
    await this.drizzle
      .update(sessions)
      .set({ 
        status,
        updatedAt: new Date(),
      })
      .where(eq(sessions.id, sessionId));
  }

  close(): void {
    this.db.close();
  }
}