package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// handleGetRaces returns all available character races
func handleGetRaces(c *fiber.Ctx) error {
	races := dmAssistant.GetAllRaces()
	return c.JSON(fiber.Map{
		"races": races,
	})
}

// handleGetClasses returns all available character classes
func handleGetClasses(c *fiber.Ctx) error {
	classes := dmAssistant.GetAllClasses()
	return c.JSON(fiber.Map{
		"classes": classes,
	})
}

// handleCreateCharacter creates a new character based on player choices
func handleCreateCharacter(c *fiber.Ctx) error {
	var req CharacterCreationRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse character creation request: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate required fields
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Character name is required"})
	}
	if req.RaceID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Race selection is required"})
	}
	if req.ClassID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Class selection is required"})
	}

	// Get race and class
	race := dmAssistant.GetRaceByID(req.RaceID)
	if race == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid race ID"})
	}

	class := dmAssistant.GetClassByID(req.ClassID)
	if class == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid class ID"})
	}

	// Build character
	character := buildCharacter(req, race, class)

	// Generate backstory if not provided
	backstory := req.Backstory
	if backstory == "" {
		generatedBackstory, err := dmAssistant.GenerateCharacterBackstory(req.Name, race, class)
		if err != nil {
			log.Printf("Failed to generate backstory: %v", err)
		} else {
			backstory = generatedBackstory
		}
	}

	return c.JSON(CharacterCreationResponse{
		Character: character,
		Message: backstory,
	})
}

// handleDMChat handles chat interactions with the AI DM
func handleDMChat(c *fiber.Ctx) error {
	var req DMChatRequest

	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse DM chat request: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if len(req.Messages) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No messages provided"})
	}

	// Get response from DM assistant
	response, err := dmAssistant.ChatWithDM(req)
	if err != nil {
		log.Printf("DM chat failed: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "DM is temporarily unavailable"})
	}

	return c.JSON(response)
}

// handleSuggestBuild suggests a character build based on preferences
func handleSuggestBuild(c *fiber.Ctx) error {
	var req struct {
		Preferences string `json:"preferences"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse build suggestion request: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Preferences == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Please provide your preferences"})
	}

	response, err := dmAssistant.SuggestCharacterBuild(req.Preferences)
	if err != nil {
		log.Printf("Build suggestion failed: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "DM is temporarily unavailable"})
	}

	return c.JSON(response)
}

// handleGenerateBackstory generates a backstory for a character
func handleGenerateBackstory(c *fiber.Ctx) error {
	var req struct {
		Name    string `json:"name"`
		RaceID  string `json:"raceId"`
		ClassID string `json:"classId"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse backstory generation request: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name == "" || req.RaceID == "" || req.ClassID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name, race, and class are required"})
	}

	race := dmAssistant.GetRaceByID(req.RaceID)
	if race == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid race ID"})
	}

	class := dmAssistant.GetClassByID(req.ClassID)
	if class == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid class ID"})
	}

	backstory, err := dmAssistant.GenerateCharacterBackstory(req.Name, race, class)
	if err != nil {
		log.Printf("Backstory generation failed: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate backstory"})
	}

	return c.JSON(fiber.Map{
		"backstory": backstory,
	})
}

// buildCharacter constructs a Character from creation request and templates
func buildCharacter(req CharacterCreationRequest, race *Race, class *Class) Character {
	// Start with class base stats
	stats := class.BaseStats

	// Apply race bonuses
	if bonus, ok := race.StatBonuses["hp"]; ok {
		stats.HP += bonus
		stats.MaxHP += bonus
	}
	if bonus, ok := race.StatBonuses["attack"]; ok {
		stats.Attack += bonus
	}
	if bonus, ok := race.StatBonuses["defense"]; ok {
		stats.Defense += bonus
	}
	if bonus, ok := race.StatBonuses["speed"]; ok {
		stats.Speed += bonus
	}

	// Apply class bonuses
	if bonus, ok := class.StatBonuses["hp"]; ok {
		stats.HP += bonus
		stats.MaxHP += bonus
	}
	if bonus, ok := class.StatBonuses["attack"]; ok {
		stats.Attack += bonus
	}
	if bonus, ok := class.StatBonuses["defense"]; ok {
		stats.Defense += bonus
	}
	if bonus, ok := class.StatBonuses["speed"]; ok {
		stats.Speed += bonus
	}

	// Override with custom stats if provided
	if req.Stats != nil {
		stats = *req.Stats
	}

	// Set position
	position := Position{X: 0, Y: 0}
	if req.Position != nil {
		position = *req.Position
	}

	// Build weapons from templates
	weapons := []Weapon{}
	for _, wt := range class.Weapons {
		weapons = append(weapons, Weapon{
			ID:       NewID(),
			Name:     wt.Name,
			Damage:   wt.Damage,
			Accuracy: wt.Accuracy,
		})
	}

	// Build abilities from race and class templates
	abilities := []Ability{}
	for _, at := range race.Abilities {
		abilities = append(abilities, Ability{
			ID:       NewID(),
			Name:     at.Name,
			Cooldown: at.Cooldown,
			Effect:   at.Effect,
			Power:    at.Power,
		})
	}
	for _, at := range class.Abilities {
		abilities = append(abilities, Ability{
			ID:       NewID(),
			Name:     at.Name,
			Cooldown: at.Cooldown,
			Effect:   at.Effect,
			Power:    at.Power,
		})
	}

	// Build items from templates
	items := []Item{}
	for _, it := range class.Items {
		items = append(items, Item{
			ID:     NewID(),
			Name:   it.Name,
			Type:   it.Type,
			Effect: it.Effect,
		})
	}

	return Character{
		ID:               NewID(),
		Name:             req.Name,
		Stats:            stats,
		Position:         position,
		Weapons:          weapons,
		Abilities:        abilities,
		Items:            items,
		AbilityCooldowns: make(map[string]int),
		IsPlayer:         true,
	}
}
