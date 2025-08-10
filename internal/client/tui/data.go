package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// initDataView инициализирует просмотр данных
func initDataView(m Model) Model {
	m.currentState = "view_data"

	// Заглушка с тестовыми данными
	/*
		items := []list.Item{
			Credential{Type: "Website", Username: "user1", Metadata: "example.com"},
			Credential{Type: "Bank Card", Metadata: "VISA •••• 1234"},
		}
	*/

	// Здесь будет инициализация списка данных
	return m
}

// updateViewData обрабатывает сообщения при просмотре данных
func updateViewData(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.currentState = "menu"
		}
	}
	return m, nil
}

// renderViewData отображает список данных
func renderViewData(m Model) string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render("Ваши данные"),
		"• Website: example.com (user1)\n• Bank Card: VISA •••• 1234",
		normalStyle.Render("Esc: назад в меню"),
	)
}
