package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// initLoginForm инициализирует форму входа
func initLoginForm(m Model) Model {
	m.currentState = "login"
	m.inputs = make([]textinput.Model, 2)

	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Логин"
	m.inputs[0].Focus()
	m.inputs[0].Width = 30

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Пароль"
	m.inputs[1].EchoMode = textinput.EchoPassword
	m.inputs[1].Width = 30

	m.focusedInput = 0
	return m
}

// updateLogin обрабатывает сообщения в форме входа
func updateLogin(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.currentState = "menu"
			m.inputs = nil
		case "enter":
			if m.focusedInput == 0 {
				m.focusedInput = 1
			} else {
				// Заглушка для логики входа
				m.user = &User{
					Username: m.inputs[0].Value(),
					Token:    "fake-token",
				}
				m.currentState = "view_data"
			}
		case "tab", "shift+tab":
			m.focusedInput = 1 - m.focusedInput // Переключаем между 0 и 1
		}
	}

	// Обновляем активное поле ввода
	var cmd tea.Cmd
	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
	return m, cmd
}

// renderLogin отображает форму входа
func renderLogin(m Model) string {
	s := titleStyle.Render("Вход в систему\n\n")
	s += fmt.Sprintf("%s\n%s\n\n",
		m.inputs[0].View(),
		m.inputs[1].View())
	s += normalStyle.Render("Tab: переключение • Enter: подтвердить • Esc: назад")
	return s
}
