package main

import (
	"fmt"
	"os"
	"strings"
)

// ScenarioService handles scenario management
type ScenarioService struct {
	scenarioDir string
}

// NewScenarioService creates a new scenario service
func NewScenarioService(scenarioDir string) *ScenarioService {
	return &ScenarioService{
		scenarioDir: scenarioDir,
	}
}

// GetAvailableScenarios returns a list of available scenario files
func (s *ScenarioService) GetAvailableScenarios() ([]string, error) {
	files, err := os.ReadDir(s.scenarioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenarios directory: %w", err)
	}

	var scenarios []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			scenarios = append(scenarios, strings.TrimSuffix(file.Name(), ".yaml"))
		}
	}

	return scenarios, nil
}
