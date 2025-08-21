package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// buildVersion, buildDate, buildCommit устанавливаются при сборке через ldflags.
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func updateAbout(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc":
			m.currentState = "menu"
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func renderAbout(m Model) string {
	var b strings.Builder

	title := titleStyle.
		Width(50).
		Align(lipgloss.Center).
		Render("О программе")
	b.WriteString(title + "\n\n")

	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	info := fmt.Sprintf("Версия: %s\nДата сборки: %s\nКоммит сборки: %s", version, date, commit)
	infoStyle := lipgloss.NewStyle().
		Width(50).
		PaddingLeft(4).
		Align(lipgloss.Left)
	b.WriteString(infoStyle.Render(info) + "\n\n")

	hint := hintStyle.
		Width(50).
		Align(lipgloss.Center).
		Render("Enter/Esc: назад • Ctrl+C: выход")
	b.WriteString(hint)

	return b.String()
}
