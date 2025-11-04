package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#0EA5E9") // Cyan
	ColorSuccess   = lipgloss.Color("#10B981") // Green
	ColorDanger    = lipgloss.Color("#EF4444") // Red
	ColorWarning   = lipgloss.Color("#F59E0B") // Yellow
	ColorInfo      = lipgloss.Color("#3B82F6") // Blue

	// UI colors
	ColorBorder     = lipgloss.Color("#4B5563")
	ColorText       = lipgloss.Color("#F3F4F6")
	ColorTextMuted  = lipgloss.Color("#9CA3AF")
	ColorBackground = lipgloss.Color("#1F2937")
)

// Base styles
var (
	// Container styles
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	// Panel styles
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2).
			MarginBottom(1)

	PlayerPanelStyle = PanelStyle.Copy().
				BorderForeground(ColorSuccess)

	EnemyPanelStyle = PanelStyle.Copy().
				BorderForeground(ColorDanger)

	// Character name styles
	PlayerNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	EnemyNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDanger)

	// Stat styles
	StatLabelStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	StatValueStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)

	// Health bar styles
	HealthBarStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	HealthBarLowStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	HealthBarCriticalStyle = lipgloss.NewStyle().
				Foreground(ColorDanger)

	// Menu styles
	MenuItemStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorText)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Foreground(ColorPrimary).
				Bold(true).
				Background(lipgloss.Color("#374151"))

	// Combat log styles
	LogStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Italic(true).
			Padding(0, 1)

	// Round indicator style
	RoundStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			Align(lipgloss.Center).
			Padding(0, 2)

	// Turn indicator style
	TurnIndicatorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWarning).
				Background(lipgloss.Color("#78350F")).
				Padding(0, 1).
				MarginLeft(1)

	// Status message styles
	SuccessMessageStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(ColorDanger).
				Bold(true)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Italic(true)

	// Banner style
	BannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Align(lipgloss.Center).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 4).
			MarginBottom(1)

	// Game over styles
	VictoryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess).
			Align(lipgloss.Center).
			Padding(1, 2)

	DefeatStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDanger).
			Align(lipgloss.Center).
			Padding(1, 2)
)

// RenderHealthBar renders a health bar with appropriate styling
func RenderHealthBar(hp, maxHP int, width int) string {
	if maxHP == 0 {
		return ""
	}

	percent := float64(hp) / float64(maxHP)
	filled := int(percent * float64(width))

	// Choose style based on health percentage
	var style lipgloss.Style
	if percent < 0.3 {
		style = HealthBarCriticalStyle
	} else if percent < 0.6 {
		style = HealthBarLowStyle
	} else {
		style = HealthBarStyle
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return style.Render(bar)
}

// RenderDivider renders a divider line
func RenderDivider(width int, char string) string {
	divider := ""
	for i := 0; i < width; i++ {
		divider += char
	}
	return lipgloss.NewStyle().Foreground(ColorBorder).Render(divider)
}
