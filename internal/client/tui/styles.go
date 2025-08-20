package tui

import "github.com/charmbracelet/lipgloss"

// Стили интерфейса (для заголовков, ошибок, активных/неактивных полей и подсказок).
var (
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	errorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	selectedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	normalStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	hintStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	activeFieldStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	inactiveFieldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	readonlyStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
	// Стиль для блока с полем формы
	formBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("15")).
			Padding(1, 1)
)
