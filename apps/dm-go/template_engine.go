package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

// TemplateEngine handles HTML template rendering
type TemplateEngine struct {
	templates *template.Template
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() (*TemplateEngine, error) {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"formatHealth":   formatHealth,
		"getHealthColor": getHealthColor,
		"formatPosition": formatPosition,
		"isPlayerTurn": func(state State) bool {
			return isPlayerTurn(state)
		},
		"renderCharacterClass": renderCharacterClass,
		"renderHealthBar":      renderHealthBar,
		"characterAt": func(characters []Character, x, y int) *Character {
			for i := range characters {
				if characters[i].Position.X == x && characters[i].Position.Y == y {
					return &characters[i]
				}
			}
			return nil
		},
		"iterate": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"percentHealth": func(hp, maxHp int) int {
			if maxHp == 0 {
				return 0
			}
			return (hp * 100) / maxHp
		},
		"json": func(v interface{}) string {
			// Simple JSON encoding for template
			data, _ := json.Marshal(v)
			return string(data)
		},
	}).ParseFS(templatesFS, "templates/*.html")

	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &TemplateEngine{templates: tmpl}, nil
}

// RenderGamePage renders the main game page
func (te *TemplateEngine) RenderGamePage(state State, sessionID string, isPlayerTurn bool) (string, error) {
	data := struct {
		State        State
		SessionID    string
		IsPlayerTurn bool
		CurrentChar  *Character
	}{
		State:        state,
		SessionID:    sessionID,
		IsPlayerTurn: isPlayerTurn,
		CurrentChar:  GetCurrentCharacter(state),
	}

	var buf bytes.Buffer
	err := te.templates.ExecuteTemplate(&buf, "game.html", data)
	if err != nil {
		return "", fmt.Errorf("failed to execute game template: %w", err)
	}

	return buf.String(), nil
}

// RenderHomePage renders the home page
func (te *TemplateEngine) RenderHomePage() (string, error) {
	var buf bytes.Buffer
	err := te.templates.ExecuteTemplate(&buf, "home.html", nil)
	if err != nil {
		return "", fmt.Errorf("failed to execute home template: %w", err)
	}

	return buf.String(), nil
}

// RenderScenariosPage renders the scenarios selection page
func (te *TemplateEngine) RenderScenariosPage(scenarios []string) (string, error) {
	data := struct {
		Scenarios []ScenarioData
	}{
		Scenarios: make([]ScenarioData, len(scenarios)),
	}

	for i, name := range scenarios {
		data.Scenarios[i] = ScenarioData{
			Name:        name,
			DisplayName: formatScenarioName(name),
		}
	}

	var buf bytes.Buffer
	err := te.templates.ExecuteTemplate(&buf, "scenarios.html", data)
	if err != nil {
		return "", fmt.Errorf("failed to execute scenarios template: %w", err)
	}

	return buf.String(), nil
}

// Template helper functions
func formatHealth(hp, maxHp int) string {
	return fmt.Sprintf("%d/%d", hp, maxHp)
}

func getHealthColor(hp, maxHp int) string {
	if hp == 0 {
		return "#dc3545" // Red for dead
	}
	if hp < maxHp/3 {
		return "#ffc107" // Yellow for low health
	}
	return "#28a745" // Green for healthy
}

func formatPosition(pos Position) string {
	return fmt.Sprintf("(%d, %d)", pos.X, pos.Y)
}

func isPlayerTurn(state State) bool {
	currentChar := GetCurrentCharacter(state)
	return currentChar != nil && currentChar.IsPlayer
}

func renderCharacterClass(char Character, isCurrent bool) string {
	classes := []string{"character"}

	if char.IsPlayer {
		classes = append(classes, "player")
	} else {
		classes = append(classes, "enemy")
	}

	if isCurrent {
		classes = append(classes, "current-turn")
	}

	if char.Stats.HP == 0 {
		classes = append(classes, "dead")
	} else if char.Stats.HP < char.Stats.MaxHP/3 {
		classes = append(classes, "low-health")
	}

	return strings.Join(classes, " ")
}

func renderHealthBar(hp, maxHp int) string {
	percentage := (hp * 100) / maxHp
	color := getHealthColor(hp, maxHp)

	return fmt.Sprintf(`<div class="health-bar">
		<div class="health-fill" style="width: %d%%; background-color: %s;"></div>
	</div>`, percentage, color)
}

func formatScenarioName(name string) string {
	return strings.Title(strings.ReplaceAll(name, "-", " "))
}

// ScenarioData for template rendering
type ScenarioData struct {
	Name        string
	DisplayName string
}
