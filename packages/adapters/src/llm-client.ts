import OpenAI from 'openai';
import type { State } from '@smol-dungeon/schema';

export interface LLMConfig {
  baseURL?: string;
  apiKey?: string;
  model: string;
  maxTokens?: number;
  temperature?: number;
}

export class LLMClient {
  private client: OpenAI;
  private config: LLMConfig;

  constructor(config: LLMConfig) {
    this.config = config;
    this.client = new OpenAI({
      baseURL: config.baseURL,
      apiKey: config.apiKey || 'dummy-key',
    });
  }

  async generateNarration(
    state: State,
    events: string[],
    context: string = ''
  ): Promise<string> {
    const systemPrompt = `You are a dungeon master narrating a combat encounter. 
Keep narration concise, dramatic, and focused on the action. 
Describe what happens without making decisions for the players.`;

    const userPrompt = `
Current situation:
${context}

Recent events:
${events.join('\n')}

Current state:
Round ${state.round}
Players: ${state.characters.filter(c => c.isPlayer).map(c => `${c.name} (${c.stats.hp}/${c.stats.maxHp} HP)`).join(', ')}
Enemies: ${state.characters.filter(c => !c.isPlayer).map(c => `${c.name} (${c.stats.hp}/${c.stats.maxHp} HP)`).join(', ')}

Provide a brief, vivid narration of what just happened:`;

    try {
      const response = await this.client.chat.completions.create({
        model: this.config.model,
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: userPrompt },
        ],
        max_tokens: this.config.maxTokens || 150,
        temperature: this.config.temperature || 0.7,
      });

      return response.choices[0]?.message?.content?.trim() || 'The battle continues...';
    } catch (error) {
      console.error('LLM narration failed:', error);
      return 'The battle continues...';
    }
  }

  async suggestEnemyAction(
    state: State,
    enemyId: string,
    context: string = ''
  ): Promise<string> {
    const enemy = state.characters.find(c => c.id === enemyId && !c.isPlayer);
    if (!enemy) {
      throw new Error(`Enemy not found: ${enemyId}`);
    }

    const systemPrompt = `You are controlling an enemy in combat. 
Choose the most tactically sound action based on the current situation.
Respond with only the action type: "Attack", "Defend", "Ability", "UseItem", or "Flee".`;

    const userPrompt = `
Enemy: ${enemy.name}
HP: ${enemy.stats.hp}/${enemy.stats.maxHp}
Available weapons: ${enemy.weapons.map(w => w.name).join(', ')}
Available abilities: ${enemy.abilities.map(a => a.name).join(', ')}
Available items: ${enemy.items.map(i => i.name).join(', ')}

Targets:
${state.characters.filter(c => c.isPlayer && c.stats.hp > 0).map(c => `${c.name} (${c.stats.hp}/${c.stats.maxHp} HP)`).join(', ')}

Context: ${context}

What should ${enemy.name} do?`;

    try {
      const response = await this.client.chat.completions.create({
        model: this.config.model,
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: userPrompt },
        ],
        max_tokens: 50,
        temperature: 0.3,
      });

      const action = response.choices[0]?.message?.content?.trim() || 'Attack';
      return action;
    } catch (error) {
      console.error('LLM action suggestion failed:', error);
      return 'Attack';
    }
  }
}