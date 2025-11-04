package saves

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"smoldungeon-cli/engine"
)

// GameSave represents a saved game
type GameSave struct {
	Name         string         `json:"name"`
	ScenarioName string         `json:"scenarioName"`
	State        engine.State   `json:"state"`
	Logs         []string       `json:"logs"`
	SavedAt      time.Time      `json:"savedAt"`
}

// SaveGame saves the current game state
func SaveGame(saveDir, name, scenarioName string, state engine.State, logs []string) error {
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	save := GameSave{
		Name:         name,
		ScenarioName: scenarioName,
		State:        state,
		Logs:         logs,
		SavedAt:      time.Now(),
	}

	data, err := json.MarshalIndent(save, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	savePath := filepath.Join(saveDir, name+".json")
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	return nil
}

// LoadGame loads a saved game
func LoadGame(saveDir, name string) (*GameSave, error) {
	savePath := filepath.Join(saveDir, name+".json")
	data, err := os.ReadFile(savePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, fmt.Errorf("failed to parse save file: %w", err)
	}

	return &save, nil
}

// SaveInfo contains metadata about a save file
type SaveInfo struct {
	Name         string
	ScenarioName string
	Round        int
	SavedAt      time.Time
	IsComplete   bool
}

// ListSaves returns information about all save files
func ListSaves(saveDir string) ([]SaveInfo, error) {
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		return []SaveInfo{}, nil
	}

	entries, err := os.ReadDir(saveDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}

	var saves []SaveInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := entry.Name()[:len(entry.Name())-5] // Remove .json
		save, err := LoadGame(saveDir, name)
		if err != nil {
			continue // Skip corrupted saves
		}

		saves = append(saves, SaveInfo{
			Name:         name,
			ScenarioName: save.ScenarioName,
			Round:        save.State.Round,
			SavedAt:      save.SavedAt,
			IsComplete:   save.State.IsComplete,
		})
	}

	// Sort by save time (most recent first)
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].SavedAt.After(saves[j].SavedAt)
	})

	return saves, nil
}

// DeleteSave deletes a save file
func DeleteSave(saveDir, name string) error {
	savePath := filepath.Join(saveDir, name+".json")
	if err := os.Remove(savePath); err != nil {
		return fmt.Errorf("failed to delete save file: %w", err)
	}
	return nil
}

// AutoSaveName generates an auto-save name
func AutoSaveName(scenarioName string) string {
	return fmt.Sprintf("auto_%s", scenarioName)
}
