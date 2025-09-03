import { useState, useEffect } from 'react';

// Simplified types for the web client
interface Character {
  id: string;
  name: string;
  stats: {
    hp: number;
    maxHp: number;
    attack: number;
    defense: number;
    speed: number;
  };
  position: { x: number; y: number };
  weapons: Array<{ id: string; name: string; damage: number; accuracy: number }>;
  abilities: Array<{ id: string; name: string; cooldown: number; effect: string; power: number }>;
  items: Array<{ id: string; name: string; type: string; effect: string }>;
  abilityCooldowns: Record<string, number>;
  isPlayer: boolean;
}

interface State {
  round: number;
  characters: Character[];
  turnOrder: string[];
  currentTurn: number;
  isComplete: boolean;
  winner?: string;
}

interface Action {
  kind: 'Attack' | 'Defend' | 'Ability' | 'UseItem' | 'Flee';
  attacker?: string;
  target?: string;
  weapon?: string;
  actor?: string;
  ability?: string;
  item?: string;
}

interface GameProps {
  seed?: number;
}

export function Game({ seed = Date.now() }: GameProps): JSX.Element {
  const [message, setMessage] = useState<string>('Welcome to SmolDungeon! The Go-powered DM server is running.');
  const [serverStatus, setServerStatus] = useState<string>('Checking...');

  useEffect(() => {
    checkServerStatus();
  }, []);

  const checkServerStatus = async (): Promise<void> => {
    try {
      const response = await fetch('http://localhost:3000/health');
      if (response.ok) {
        const data = await response.json();
        setServerStatus(`✅ Go DM Server is running (Port 3000)`);
      } else {
        setServerStatus('❌ Go DM Server not responding');
      }
    } catch (error) {
      setServerStatus('❌ Cannot connect to Go DM Server');
    }
  };

  return (
    <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif', maxWidth: '800px', margin: '0 auto' }}>
      <h1 style={{ color: '#007bff', textAlign: 'center' }}>🚀 SmolDungeon - Go Powered</h1>

      <div style={{ background: '#f8f9fa', padding: '20px', borderRadius: '8px', margin: '20px 0' }}>
        <h2>Server Status</h2>
        <p style={{ fontSize: '18px', fontWeight: 'bold' }}>{serverStatus}</p>
        <button
          onClick={checkServerStatus}
          style={{
            padding: '10px 20px',
            background: '#007bff',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            marginTop: '10px'
          }}
        >
          Refresh Status
        </button>
      </div>

      <div style={{ background: '#e9ecef', padding: '20px', borderRadius: '8px', margin: '20px 0' }}>
        <h2>🎯 Project Status</h2>
        <ul>
          <li>✅ <strong>Go DM Server:</strong> High-performance Fiber-based backend</li>
          <li>✅ <strong>REST API:</strong> Full combat mechanics and session management</li>
          <li>✅ <strong>Database:</strong> SQLite with event sourcing</li>
          <li>✅ <strong>LLM Integration:</strong> OpenAI-compatible AI features</li>
          <li>🔄 <strong>Web Client:</strong> Simplified interface (in progress)</li>
        </ul>
      </div>

      <div style={{ background: '#d4edda', padding: '20px', borderRadius: '8px', margin: '20px 0' }}>
        <h2>🎮 API Endpoints</h2>
        <p>Test the Go DM server directly:</p>
        <div style={{ fontFamily: 'monospace', background: 'white', padding: '10px', borderRadius: '4px' }}>
          GET  http://localhost:3000/health<br/>
          GET  http://localhost:3000/sessions<br/>
          POST http://localhost:3000/tools/get_state_summary<br/>
          POST http://localhost:3000/tools/roll_check<br/>
          POST http://localhost:3000/tools/apply_action
        </div>
      </div>

      <div style={{ textAlign: 'center', marginTop: '40px', color: '#666' }}>
        <p>🚀 <strong>Clean, Fast, Go-powered SmolDungeon</strong> 🚀</p>
        <p>The future of turn-based combat gaming</p>
      </div>
    </div>
  );
}