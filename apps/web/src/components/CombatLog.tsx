interface CombatLogProps {
  logs: string[];
  maxLines?: number;
}

export function CombatLog({ logs, maxLines = 10 }: CombatLogProps): JSX.Element {
  const recentLogs = logs.slice(-maxLines);

  return (
    <div style={{
      border: '1px solid #ccc',
      borderRadius: '4px',
      padding: '10px',
      backgroundColor: '#f8f9fa',
      height: '200px',
      overflowY: 'auto'
    }}>
      <h3 style={{ color: '#6f42c1', margin: '0 0 10px 0' }}>Combat Log</h3>
      <div style={{ height: `${maxLines * 1.4}em`, overflowY: 'auto' }}>
        {recentLogs.map((log, index) => (
          <div key={index} style={{ marginBottom: '2px', color: '#333' }}>
            {log}
          </div>
        ))}
      </div>
    </div>
  );
}