package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// DMAssistant provides AI-powered DM assistance for character creation and gameplay
type DMAssistant struct {
	llmClient *LLMClient
	races     []Race
	classes   []Class
}

// NewDMAssistant creates a new DM assistant
func NewDMAssistant(llmClient *LLMClient) *DMAssistant {
	return &DMAssistant{
		llmClient: llmClient,
		races:     getAvailableRaces(),
		classes:   getAvailableClasses(),
	}
}

// ChatWithDM handles conversational interaction with the AI DM
func (dm *DMAssistant) ChatWithDM(req DMChatRequest) (DMChatResponse, error) {
	systemPrompt := dm.buildSystemPrompt(req.Context)

	// Convert DMChatMessages to OpenAI format
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	for _, msg := range req.Messages {
		var role string
		switch msg.Role {
		case "user":
			role = openai.ChatMessageRoleUser
		case "assistant":
			role = openai.ChatMessageRoleAssistant
		case "system":
			role = openai.ChatMessageRoleSystem
		default:
			role = openai.ChatMessageRoleUser
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	resp, err := dm.llmClient.remoteClient.CreateChatCompletion(
		nil,
		openai.ChatCompletionRequest{
			Model:       dm.llmClient.config.Model,
			Messages:    messages,
			MaxTokens:   500,
			Temperature: 0.7,
		},
	)

	if err != nil {
		log.Printf("DM chat failed: %v", err)
		return DMChatResponse{
			Message: "The DM seems distracted. Please try again.",
		}, err
	}

	if len(resp.Choices) == 0 {
		return DMChatResponse{
			Message: "The DM seems distracted. Please try again.",
		}, fmt.Errorf("no response from LLM")
	}

	message := resp.Choices[0].Message.Content
	suggestions := dm.extractSuggestions(message, req.Context)

	return DMChatResponse{
		Message:     message,
		Suggestions: suggestions,
	}, nil
}

// GenerateCharacterBackstory generates a backstory for a character
func (dm *DMAssistant) GenerateCharacterBackstory(name string, race *Race, class *Class) (string, error) {
	prompt := fmt.Sprintf(`Generate a compelling backstory for a D&D character with these details:
- Name: %s
- Race: %s (%s)
- Class: %s (%s)

Create a 2-3 paragraph backstory that includes:
- Their origin and upbringing
- What led them to become a %s
- A personal motivation or goal
- A hint of conflict or challenge in their past

Keep it engaging and suitable for a fantasy adventure.`,
		name,
		race.Name, race.Description,
		class.Name, class.Description,
		class.Name,
	)

	resp, err := dm.llmClient.remoteClient.CreateChatCompletion(
		nil,
		openai.ChatCompletionRequest{
			Model: dm.llmClient.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a creative dungeon master writing character backstories.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   400,
			Temperature: 0.8,
		},
	)

	if err != nil {
		log.Printf("Backstory generation failed: %v", err)
		return fmt.Sprintf("%s is a %s %s seeking adventure and glory.", name, race.Name, class.Name), nil
	}

	if len(resp.Choices) == 0 {
		return fmt.Sprintf("%s is a %s %s seeking adventure and glory.", name, race.Name, class.Name), nil
	}

	return resp.Choices[0].Message.Content, nil
}

// SuggestCharacterBuild suggests a character build based on player preferences
func (dm *DMAssistant) SuggestCharacterBuild(preferences string) (DMChatResponse, error) {
	racesJSON, _ := json.MarshalIndent(dm.races, "", "  ")
	classesJSON, _ := json.MarshalIndent(dm.classes, "", "  ")

	prompt := fmt.Sprintf(`Player preferences: %s

Available races:
%s

Available classes:
%s

Based on the player's preferences, suggest a race and class combination that would work well. Explain why this combination fits their preferences and what playstyle they should expect.`,
		preferences,
		string(racesJSON),
		string(classesJSON),
	)

	resp, err := dm.llmClient.remoteClient.CreateChatCompletion(
		nil,
		openai.ChatCompletionRequest{
			Model: dm.llmClient.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful dungeon master assisting with character creation. Provide clear, enthusiastic guidance.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   400,
			Temperature: 0.7,
		},
	)

	if err != nil {
		log.Printf("Build suggestion failed: %v", err)
		return DMChatResponse{
			Message: "I'd suggest starting with a Human Fighter - it's a solid, versatile choice for any adventurer!",
		}, err
	}

	if len(resp.Choices) == 0 {
		return DMChatResponse{
			Message: "I'd suggest starting with a Human Fighter - it's a solid, versatile choice for any adventurer!",
		}, fmt.Errorf("no response from LLM")
	}

	return DMChatResponse{
		Message:     resp.Choices[0].Message.Content,
		Suggestions: []string{"Continue with this build", "Try a different combination"},
	}, nil
}

// buildSystemPrompt creates the system prompt based on context
func (dm *DMAssistant) buildSystemPrompt(context string) string {
	basePrompt := `You are an experienced and enthusiastic Dungeon Master helping players create their characters.

Your role is to:
- Guide players through character creation with patience and enthusiasm
- Explain D&D mechanics in simple, accessible terms
- Suggest builds and options that match player preferences
- Generate creative backstories and character concepts
- Provide tactical advice about different choices
- Keep the tone friendly, encouraging, and immersive

When discussing game mechanics:
- Stats: HP (health), Attack (damage), Defense (armor), Speed (initiative)
- Combat uses D20 rolls for attacks (DC 10 to hit)
- Damage is affected by weapon damage, attack stat, and D6 variance
- Abilities have cooldowns and special effects

Be concise but engaging. Use vivid language and D&D terminology to make the experience immersive.`

	if context != "" {
		basePrompt += fmt.Sprintf("\n\nCurrent context: %s", context)
	}

	// Add available races and classes to the system prompt
	racesJSON, _ := json.MarshalIndent(dm.races, "", "  ")
	classesJSON, _ := json.MarshalIndent(dm.classes, "", "  ")

	basePrompt += fmt.Sprintf(`

Available races:
%s

Available classes:
%s`, string(racesJSON), string(classesJSON))

	return basePrompt
}

// extractSuggestions extracts action suggestions from the DM's response
func (dm *DMAssistant) extractSuggestions(message string, context string) []string {
	// Based on context, provide relevant suggestions
	suggestions := []string{}

	lowerMsg := strings.ToLower(message)
	lowerContext := strings.ToLower(context)

	if strings.Contains(lowerContext, "race") {
		if strings.Contains(lowerMsg, "human") {
			suggestions = append(suggestions, "Choose Human")
		}
		if strings.Contains(lowerMsg, "elf") {
			suggestions = append(suggestions, "Choose Elf")
		}
		if strings.Contains(lowerMsg, "dwarf") {
			suggestions = append(suggestions, "Choose Dwarf")
		}
		if strings.Contains(lowerMsg, "halfling") {
			suggestions = append(suggestions, "Choose Halfling")
		}
		suggestions = append(suggestions, "See all races", "Ask for recommendation")
	} else if strings.Contains(lowerContext, "class") {
		if strings.Contains(lowerMsg, "fighter") {
			suggestions = append(suggestions, "Choose Fighter")
		}
		if strings.Contains(lowerMsg, "wizard") {
			suggestions = append(suggestions, "Choose Wizard")
		}
		if strings.Contains(lowerMsg, "rogue") {
			suggestions = append(suggestions, "Choose Rogue")
		}
		if strings.Contains(lowerMsg, "cleric") {
			suggestions = append(suggestions, "Choose Cleric")
		}
		suggestions = append(suggestions, "See all classes", "Ask for recommendation")
	} else if strings.Contains(lowerContext, "backstory") {
		suggestions = append(suggestions, "Generate backstory", "Write my own", "Skip for now")
	}

	return suggestions
}

// getAvailableRaces returns the list of available character races
func getAvailableRaces() []Race {
	return []Race{
		{
			ID:          "human",
			Name:        "Human",
			Description: "Versatile and adaptable, humans are found in every corner of the world.",
			StatBonuses: map[string]int{"hp": 2, "attack": 1, "defense": 1},
			Abilities: []AbilityTemplate{
				{Name: "Determination", Cooldown: 5, Effect: "heal", Power: 10},
			},
		},
		{
			ID:          "elf",
			Name:        "Elf",
			Description: "Graceful and long-lived, elves are known for their agility and magical affinity.",
			StatBonuses: map[string]int{"speed": 3, "attack": 2},
			Abilities: []AbilityTemplate{
				{Name: "Keen Senses", Cooldown: 3, Effect: "buff", Power: 5},
			},
		},
		{
			ID:          "dwarf",
			Name:        "Dwarf",
			Description: "Sturdy and resilient, dwarves are master craftsmen and fierce warriors.",
			StatBonuses: map[string]int{"hp": 4, "defense": 3},
			Abilities: []AbilityTemplate{
				{Name: "Stone Resilience", Cooldown: 4, Effect: "buff", Power: 8},
			},
		},
		{
			ID:          "halfling",
			Name:        "Halfling",
			Description: "Small but brave, halflings are lucky and surprisingly tough.",
			StatBonuses: map[string]int{"speed": 2, "defense": 2},
			Abilities: []AbilityTemplate{
				{Name: "Lucky", Cooldown: 6, Effect: "buff", Power: 10},
			},
		},
	}
}

// getAvailableClasses returns the list of available character classes
func getAvailableClasses() []Class {
	return []Class{
		{
			ID:          "fighter",
			Name:        "Fighter",
			Description: "Master of martial combat, skilled with weapons and armor.",
			StatBonuses: map[string]int{"hp": 5, "attack": 3, "defense": 2},
			BaseStats:   Stat{HP: 40, MaxHP: 40, Attack: 8, Defense: 6, Speed: 5},
			Weapons: []WeaponTemplate{
				{Name: "Longsword", Damage: 8, Accuracy: 85},
				{Name: "Battleaxe", Damage: 10, Accuracy: 80},
			},
			Abilities: []AbilityTemplate{
				{Name: "Power Attack", Cooldown: 3, Effect: "damage", Power: 15},
				{Name: "Second Wind", Cooldown: 5, Effect: "heal", Power: 20},
			},
			Items: []ItemTemplate{
				{Name: "Health Potion", Type: "consumable", Effect: "heal 20 HP"},
			},
		},
		{
			ID:          "wizard",
			Name:        "Wizard",
			Description: "Wielder of arcane magic, capable of devastating spells.",
			StatBonuses: map[string]int{"attack": 4, "speed": 2},
			BaseStats:   Stat{HP: 30, MaxHP: 30, Attack: 10, Defense: 3, Speed: 6},
			Weapons: []WeaponTemplate{
				{Name: "Staff", Damage: 5, Accuracy: 90},
			},
			Abilities: []AbilityTemplate{
				{Name: "Fireball", Cooldown: 3, Effect: "damage", Power: 18},
				{Name: "Magic Shield", Cooldown: 4, Effect: "buff", Power: 10},
				{Name: "Lightning Bolt", Cooldown: 5, Effect: "damage", Power: 22},
			},
			Items: []ItemTemplate{
				{Name: "Mana Potion", Type: "consumable", Effect: "heal 15 HP"},
			},
		},
		{
			ID:          "rogue",
			Name:        "Rogue",
			Description: "Cunning and quick, rogues strike from the shadows.",
			StatBonuses: map[string]int{"attack": 2, "speed": 4},
			BaseStats:   Stat{HP: 35, MaxHP: 35, Attack: 9, Defense: 4, Speed: 8},
			Weapons: []WeaponTemplate{
				{Name: "Daggers", Damage: 6, Accuracy: 95},
				{Name: "Shortsword", Damage: 7, Accuracy: 90},
			},
			Abilities: []AbilityTemplate{
				{Name: "Sneak Attack", Cooldown: 2, Effect: "damage", Power: 16},
				{Name: "Evasion", Cooldown: 4, Effect: "buff", Power: 12},
			},
			Items: []ItemTemplate{
				{Name: "Health Potion", Type: "consumable", Effect: "heal 20 HP"},
				{Name: "Smoke Bomb", Type: "consumable", Effect: "escape combat"},
			},
		},
		{
			ID:          "cleric",
			Name:        "Cleric",
			Description: "Divine spellcaster, healer and protector of the faithful.",
			StatBonuses: map[string]int{"hp": 3, "defense": 3, "attack": 1},
			BaseStats:   Stat{HP: 38, MaxHP: 38, Attack: 7, Defense: 5, Speed: 5},
			Weapons: []WeaponTemplate{
				{Name: "Mace", Damage: 7, Accuracy: 85},
			},
			Abilities: []AbilityTemplate{
				{Name: "Healing Word", Cooldown: 3, Effect: "heal", Power: 25},
				{Name: "Divine Strike", Cooldown: 4, Effect: "damage", Power: 14},
				{Name: "Bless", Cooldown: 5, Effect: "buff", Power: 15},
			},
			Items: []ItemTemplate{
				{Name: "Health Potion", Type: "consumable", Effect: "heal 20 HP"},
				{Name: "Holy Water", Type: "consumable", Effect: "damage undead 30"},
			},
		},
	}
}

// GetRaceByID returns a race by ID
func (dm *DMAssistant) GetRaceByID(id string) *Race {
	for i := range dm.races {
		if dm.races[i].ID == id {
			return &dm.races[i]
		}
	}
	return nil
}

// GetClassByID returns a class by ID
func (dm *DMAssistant) GetClassByID(id string) *Class {
	for i := range dm.classes {
		if dm.classes[i].ID == id {
			return &dm.classes[i]
		}
	}
	return nil
}

// GetAllRaces returns all available races
func (dm *DMAssistant) GetAllRaces() []Race {
	return dm.races
}

// GetAllClasses returns all available classes
func (dm *DMAssistant) GetAllClasses() []Class {
	return dm.classes
}
