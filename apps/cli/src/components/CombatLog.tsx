import React from 'react';
import { Box, Text } from 'ink';

interface CombatLogProps {
  logs: string[];
  maxLines?: number;
}

export function CombatLog({ logs, maxLines = 10 }: CombatLogProps): JSX.Element {
  const recentLogs = logs.slice(-maxLines);

  return (
    <Box flexDirection="column" padding={1} borderStyle="single" borderColor="gray">
      <Text bold color="magenta">Combat Log</Text>
      <Box height={maxLines} flexDirection="column">
        {recentLogs.map((log, index) => (
          <Text key={index} color="white">
            {log}
          </Text>
        ))}
      </Box>
    </Box>
  );
}