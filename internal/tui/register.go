package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var fieldLabels = []string{
	"Логин",
	"Пароль",
	"Подтвердите пароль",
}

func initRegisterForm(m Model) Model {
	m.currentState = "register"
	m.inputs = make([]textinput.Model, 3)

	// Настройка полей без стандартного промпта
	for i := range m.inputs {
		m.inputs[i] = newInputField("") // Пустой placeholder
	}

	m.inputs[0].Focus()
	m.inputs[1].EchoMode = textinput.EchoPassword
	m.inputs[2].EchoMode = textinput.EchoPassword

	m.focusedInput = 0

	return m
}

func newInputField(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = " " // Один пробел вместо ">"
	ti.Cursor.Style = lipgloss.NewStyle().Background(lipgloss.Color("15"))
	ti.Width = 30
	return ti
}

func updateRegister(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.currentState = "menu"
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab", "up", "down", "enter":
			s := msg.String()
			if s == "enter" && m.focusedInput == len(m.inputs)-1 {
				// Нажали Enter на пароле - выполняем вход
				m.currentState = "menu"
			}
			if s == "up" || s == "shift+tab" {
				m.focusedInput = (m.focusedInput - 1 + len(m.inputs)) % len(m.inputs)
			} else {
				m.focusedInput = (m.focusedInput + 1) % len(m.inputs)
			}
			for i := range m.inputs {
				if i == m.focusedInput {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
	return m, cmd
}

func renderRegister(m Model) string {
	var builder strings.Builder

	// Заголовок
	builder.WriteString(titleStyle.Render("Регистрация"))
	builder.WriteString("\n\n")

	// Поля формы
	for i, input := range m.inputs {
		// Подпись поля
		label := fieldLabels[i] + ": "
		if i == m.focusedInput {
			label = activeFieldStyle.Render(label)
		} else {
			label = inactiveFieldStyle.Render(label)
		}

		// Поле ввода (убираем стандартный промпт)
		//field := strings.Replace(input.View(), ">", " ", 1)

		builder.WriteString(label + input.View() + "\n")
	}

	// Подсказки
	builder.WriteString("\n" + hintStyle.Render(
		"Tab: переключение • Enter: подтвердить • Esc: назад • Ctrl+C: выход",
	))

	return builder.String()
}
