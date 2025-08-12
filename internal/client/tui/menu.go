package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
)

// updateMenu обрабатывает сообщения в состоянии меню
func updateMenu(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "tab":
			if m.menuCursor < len(m.menuItems)-1 {
				m.menuCursor++
			}
		case "enter":
			selected := m.menuItems[m.menuCursor].title
			switch selected {
			case "Login":
				m = initLoginForm(m)
			case "Register":
				m = initRegisterForm(m)
			case "Credentials":
				m.currentType = contracts.TypeCredentials
				m.currentState = "list"
				return m, m.loadList()
			case "Exit":
				return m, tea.Quit
			}
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// renderMenu отображает главное меню
func renderMenu(m Model) string {
	var b strings.Builder

	// Заголовок
	title := titleStyle.
		Width(50).
		Align(lipgloss.Center).
		Render("GophKeeper - Менеджер паролей")
	b.WriteString(title + "\n\n")

	// Контейнер меню
	menuStyle := lipgloss.NewStyle().
		Width(50).
		PaddingLeft(4).
		Align(lipgloss.Left)

	// Элементы меню
	var menuItems []string
	for i, item := range m.menuItems {
		if m.menuCursor == i {
			menuItems = append(menuItems,
				selectedStyle.Render("> "+item.title+" - "+item.description))
		} else {
			menuItems = append(menuItems,
				normalStyle.Render("  "+item.title+" - "+item.description))
		}
	}

	// Собираем меню
	menu := strings.Join(menuItems, "\n")
	b.WriteString(menuStyle.Render(menu) + "\n\n")

	// Подсказки
	hint := hintStyle.
		Width(50).
		Align(lipgloss.Center).
		Render("↑/↓: навигация • Enter: выбор • Ctrl+C: выход")
	b.WriteString(hint)

	return b.String()
}
