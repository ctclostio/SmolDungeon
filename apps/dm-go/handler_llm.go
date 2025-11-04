package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

// handleGenerateNarration generates narrative text for game events
func handleGenerateNarration(c *fiber.Ctx) error {
	var req struct {
		State    State    `json:"state"`
		Events   []string `json:"events"`
		Context  string   `json:"context,omitempty"`
		UseLocal bool     `json:"useLocal,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	narration, err := llmClient.GenerateNarrationWithModel(req.State, req.Events, req.Context, req.UseLocal)
	if err != nil {
		log.Printf("Narration generation failed: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate narration"})
	}

	return c.JSON(fiber.Map{
		"narration": narration,
		"model":     req.UseLocal && llmClient.config.LocalEnabled,
	})
}

// handleGenerateCombatDescription generates combat descriptions
func handleGenerateCombatDescription(c *fiber.Ctx) error {
	var req struct {
		State    State      `json:"state"`
		Action   Action     `json:"action"`
		Events   []string   `json:"events"`
		Attacker *Character `json:"attacker,omitempty"`
		Target   *Character `json:"target,omitempty"`
		UseLocal bool       `json:"useLocal,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Create enhanced context for combat description
	context := fmt.Sprintf("Combat Action: %s", req.Action.Kind)
	if req.Attacker != nil && req.Target != nil {
		context += fmt.Sprintf(" - %s attacks %s", req.Attacker.Name, req.Target.Name)
	}

	narration, err := llmClient.GenerateNarrationWithModel(req.State, req.Events, context, req.UseLocal)
	if err != nil {
		log.Printf("Combat description generation failed: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate combat description"})
	}

	return c.JSON(fiber.Map{
		"description": narration,
		"action":      req.Action.Kind,
		"model":       req.UseLocal && llmClient.config.LocalEnabled,
	})
}
