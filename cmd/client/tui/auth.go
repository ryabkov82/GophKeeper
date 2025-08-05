package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	inputs  []textinput.Model // Поля ввода
	focused int               // Какое поле активно
}

func InitialModel() Model {

	// Создаем стиль для курсора
	/*
		cursorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("205")).
			Foreground(lipgloss.Color("235"))
	*/

	inputs := make([]textinput.Model, 2)

	// Поле логина
	inputs[0].CharLimit = 50 // Лимит символов
	inputs[0].Width = 30     // Ширина поля
	inputs[0] = textinput.New()
	inputs[0].Prompt = "Логин: "
	//inputs[0].Cursor.Style = cursorStyle // Применяем стиль
	inputs[0].Cursor.Blink = true
	inputs[0].Focus()

	// Поле пароля
	inputs[1].CharLimit = 50 // Лимит символов
	inputs[1].Width = 30     // Ширина поля
	inputs[1] = textinput.New()
	inputs[1].Prompt = "Пароль: "
	//inputs[1].Cursor.Style = cursorStyle // Применяем стиль
	inputs[1].Cursor.Blink = true
	inputs[1].EchoMode = textinput.EchoPassword

	return Model{
		inputs:  inputs,
		focused: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			// Переключение между полями
			s := msg.String()
			if s == "enter" && m.focused == len(m.inputs)-1 {
				// Нажали Enter на пароле - выполняем вход
				return m, nil // Здесь будет вызов gRPC
			}

			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			if m.focused >= len(m.inputs) {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focused {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Обновляем активное поле
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Обновляем только активное поле
	old := m.inputs[m.focused]
	m.inputs[m.focused], _ = old.Update(msg)
	if m.focused == 1 { // Пароль
		m.inputs[m.focused].EchoMode = textinput.EchoPassword
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	return `
  Введите логин и пароль:

  ` + m.inputs[0].View() + `
  ` + m.inputs[1].View() + `

  (Tab/↑↓ - переключение, Enter - вход, Ctrl+C - выход)
`
}
