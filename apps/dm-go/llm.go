package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// LLMConfig holds configuration for the LLM client
type LLMConfig struct {
	// Remote model settings (OpenAI compatible)
	BaseURL     string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float32

	// Local model settings
	LocalEnabled     bool
	LocalBaseURL     string
	LocalModel       string
	LocalMaxTokens   int
	LocalTemperature float32

	// Model selection
	PreferredModel string // "remote", "local", or "auto"
}

// Local model request/response structures
type LocalChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LocalChatRequest struct {
	Model       string             `json:"model"`
	Messages    []LocalChatMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float32            `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

type LocalChatChoice struct {
	Message      LocalChatMessage `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

type LocalChatResponse struct {
	Choices []LocalChatChoice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// LLMClient handles interactions with LLM services
type LLMClient struct {
	remoteClient *openai.Client
	httpClient   *http.Client
	config       LLMConfig
}

// NewLLMClient creates a new LLM client
func NewLLMClient(config LLMConfig) *LLMClient {
	var remoteClient *openai.Client
	if config.BaseURL != "" {
		openaiConfig := openai.DefaultConfig(config.APIKey)
		openaiConfig.BaseURL = config.BaseURL
		remoteClient = openai.NewClientWithConfig(openaiConfig)
	} else {
		remoteClient = openai.NewClient(config.APIKey)
	}

	return &LLMClient{
		remoteClient: remoteClient,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// GenerateNarration generates narrative text for game events
func (llm *LLMClient) GenerateNarration(state State, events []string, context string) (string, error) {
	systemPrompt := `You are a dungeon master narrating a combat encounter.
Keep narration concise, dramatic, and focused on the action.
Describe what happens without making decisions for the players.`

	var userPrompt string
	if context != "" {
		userPrompt = fmt.Sprintf(`Current situation:
%s

Recent events:
%s

Current state:
Round %d
Players: %s
Enemies: %s

Provide a brief, vivid narration of what just happened:`,
			context,
			formatEvents(events),
			state.Round,
			formatCharacters(state.Characters, true),
			formatCharacters(state.Characters, false))
	} else {
		userPrompt = fmt.Sprintf(`Recent events:
%s

Current state:
Round %d
Players: %s
Enemies: %s

Provide a brief, vivid narration of what just happened:`,
			formatEvents(events),
			state.Round,
			formatCharacters(state.Characters, true),
			formatCharacters(state.Characters, false))
	}

	resp, err := llm.remoteClient.CreateChatCompletion(
		nil,
		openai.ChatCompletionRequest{
			Model: llm.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			MaxTokens:   llm.config.MaxTokens,
			Temperature: llm.config.Temperature,
		},
	)

	if err != nil {
		return "The battle continues...", fmt.Errorf("LLM narration failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "The battle continues...", nil
	}

	return resp.Choices[0].Message.Content, nil
}

// SuggestEnemyAction suggests an action for an enemy character
func (llm *LLMClient) SuggestEnemyAction(state State, enemyID ID, context string) (string, error) {
	enemy := GetCharacterByID(state, enemyID)
	if enemy == nil || enemy.IsPlayer {
		return "Attack", fmt.Errorf("enemy not found or is player: %s", enemyID)
	}

	systemPrompt := `You are controlling an enemy in combat.
Choose the most tactically sound action based on the current situation.
Respond with only the action type: "Attack", "Defend", "Ability", "UseItem", or "Flee".`

	userPrompt := fmt.Sprintf(`Enemy: %s
HP: %d/%d
Available weapons: %s
Available abilities: %s
Available items: %s

Targets:
%s

Context: %s

What should %s do?`,
		enemy.Name,
		enemy.Stats.HP, enemy.Stats.MaxHP,
		formatWeapons(enemy.Weapons),
		formatAbilities(enemy.Abilities),
		formatItems(enemy.Items),
		formatTargets(state.Characters),
		context,
		enemy.Name)

	resp, err := llm.remoteClient.CreateChatCompletion(
		nil,
		openai.ChatCompletionRequest{
			Model: llm.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			MaxTokens:   50,
			Temperature: 0.3,
		},
	)

	if err != nil {
		return "Attack", fmt.Errorf("LLM action suggestion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "Attack", nil
	}

	action := resp.Choices[0].Message.Content
	if action == "" {
		action = "Attack"
	}

	return action, nil
}

// Helper functions for formatting
func formatEvents(events []string) string {
	if len(events) == 0 {
		return "No recent events"
	}
	return fmt.Sprintf("%s", events)
}

func formatCharacters(characters []Character, isPlayer bool) string {
	var chars []string
	for _, char := range characters {
		if char.IsPlayer == isPlayer {
			status := fmt.Sprintf("%s (%d/%d HP)", char.Name, char.Stats.HP, char.Stats.MaxHP)
			chars = append(chars, status)
		}
	}
	if len(chars) == 0 {
		return "None"
	}
	return fmt.Sprintf("%s", chars)
}

func formatWeapons(weapons []Weapon) string {
	if len(weapons) == 0 {
		return "None"
	}
	var names []string
	for _, w := range weapons {
		names = append(names, w.Name)
	}
	return fmt.Sprintf("%s", names)
}

func formatAbilities(abilities []Ability) string {
	if len(abilities) == 0 {
		return "None"
	}
	var names []string
	for _, a := range abilities {
		names = append(names, a.Name)
	}
	return fmt.Sprintf("%s", names)
}

func formatItems(items []Item) string {
	if len(items) == 0 {
		return "None"
	}
	var names []string
	for _, i := range items {
		names = append(names, i.Name)
	}
	return fmt.Sprintf("%s", names)
}

func formatTargets(characters []Character) string {
	var targets []string
	for _, char := range characters {
		if char.IsPlayer && char.Stats.HP > 0 {
			targets = append(targets, fmt.Sprintf("%s (%d/%d HP)", char.Name, char.Stats.HP, char.Stats.MaxHP))
		}
	}
	if len(targets) == 0 {
		return "None"
	}
	return fmt.Sprintf("%s", targets)
}

// callLocalModel makes a request to a local LLM API
func (llm *LLMClient) callLocalModel(messages []LocalChatMessage) (string, error) {
	req := LocalChatRequest{
		Model:       llm.config.LocalModel,
		Messages:    messages,
		MaxTokens:   llm.config.LocalMaxTokens,
		Temperature: llm.config.LocalTemperature,
		Stream:      false,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", llm.config.LocalBaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := llm.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("local model API error: %s - %s", resp.Status, string(body))
	}

	var localResp LocalChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&localResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(localResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from local model")
	}

	return localResp.Choices[0].Message.Content, nil
}

// shouldUseLocalModel determines if we should use the local model
func (llm *LLMClient) shouldUseLocalModel() bool {
	if !llm.config.LocalEnabled {
		return false
	}

	switch llm.config.PreferredModel {
	case "local":
		return true
	case "remote":
		return false
	case "auto":
		// Try local first, fallback to remote if local fails
		return true
	default:
		return false
	}
}

// GenerateNarrationWithModel generates narrative text using the appropriate model
func (llm *LLMClient) GenerateNarrationWithModel(state State, events []string, context string, useLocal bool) (string, error) {
	systemPrompt := `You are a master dungeon master narrating an epic fantasy combat encounter.
Create vivid, immersive descriptions that bring the battle to life. Focus on:
- The intensity and drama of combat actions
- Environmental details and atmosphere
- Character emotions and physical sensations
- Strategic positioning and tactical elements
- The consequences and stakes of each action

Keep descriptions engaging but concise, maintaining the flow of combat while building tension and excitement.`

	var userPrompt string
	if context != "" {
		userPrompt = fmt.Sprintf(`Current situation:
%s

Recent events:
%s

Current state:
Round %d
Players: %s
Enemies: %s

Create a vivid, dramatic narration of what just happened in this combat encounter:`,
			context,
			formatEvents(events),
			state.Round,
			formatCharacters(state.Characters, true),
			formatCharacters(state.Characters, false))
	} else {
		userPrompt = fmt.Sprintf(`Recent events:
%s

Current state:
Round %d
Players: %s
Enemies: %s

Create a vivid, dramatic narration of what just happened in this combat encounter:`,
			formatEvents(events),
			state.Round,
			formatCharacters(state.Characters, true),
			formatCharacters(state.Characters, false))
	}

	// Try local model first if enabled
	if useLocal && llm.config.LocalEnabled {
		localMessages := []LocalChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		}

		if narration, err := llm.callLocalModel(localMessages); err == nil {
			return narration, nil
		} else if llm.config.PreferredModel != "auto" {
			return "", fmt.Errorf("local model failed: %w", err)
		}
		// If auto mode and local fails, fall back to remote
	}

	// Use remote model (OpenAI compatible)
	resp, err := llm.remoteClient.CreateChatCompletion(
		nil,
		openai.ChatCompletionRequest{
			Model: llm.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			MaxTokens:   llm.config.MaxTokens,
			Temperature: llm.config.Temperature,
		},
	)

	if err != nil {
		return "The battle rages on with intense combat!", fmt.Errorf("remote model failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "The battle rages on with intense combat!", nil
	}

	return resp.Choices[0].Message.Content, nil
}
