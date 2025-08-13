package tui

import (
	"context"
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
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
	//ti.Cursor.Style = lipgloss.NewStyle().Background(lipgloss.Color("15"))
	ti.Cursor.Style = cursorStyle
	ti.Width = 30
	return ti
}

func updateRegister(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.focusedInput == len(m.inputs)-1 {
				// Логика регистрации
				if m.inputs[1].Value() != m.inputs[2].Value() {
					m.registerErr = errors.New("пароли не совпадают")
					return m, nil
				}

				return m, tea.Batch(
					tea.Printf("Регистрируем пользователя..."),
					registerUser(m.ctx, m.authService, m.inputs[0].Value(), m.inputs[1].Value()),
				)
			}

			// Переход к следующему полю
			m.focusedInput = (m.focusedInput + 1) % len(m.inputs)
			return updateInputFocus(m), nil

		case "esc":
			m.currentState = "menu"
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.focusedInput = (m.focusedInput - 1 + len(m.inputs)) % len(m.inputs)
			} else {
				m.focusedInput = (m.focusedInput + 1) % len(m.inputs)
			}
			return updateInputFocus(m), nil
		}
	case RegisterSuccessMsg:
		m.currentState = "registerSuccess" // переход в промежуточное состояние
		m.registerErr = nil
		return m, nil

	case RegisterFailedMsg:
		m.registerErr = msg.Err
		return m, nil
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

	// Отображение ошибки
	if m.registerErr != nil {
		builder.WriteString("\n" + errorStyle.Render("Ошибка: "+m.registerErr.Error()))
	}

	// Подсказки
	builder.WriteString("\n" + hintStyle.Render(
		"Tab: переключение • Enter: подтвердить • Esc: назад • Ctrl+C: выход",
	))

	return builder.String()
}

// Команда для регистрации
func registerUser(ctx context.Context, authService contracts.AuthService, login, password string) tea.Cmd {
	return func() tea.Msg {
		err := authService.RegisterUser(ctx, login, password)
		if err != nil {
			return RegisterFailedMsg{Err: err}
		}
		return RegisterSuccessMsg{}
	}
}

// Вспомогательная функция для обновления фокуса
func updateInputFocus(m Model) Model {
	for i := range m.inputs {
		if i == m.focusedInput {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return m
}

func updateRegisterSuccess(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.currentState = "menu" // переход в меню по Enter
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func renderRegisterSuccess(m Model) string {
	return titleStyle.Render("Регистрация") + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("Успешная регистрация!") + "\n\n" +
		hintStyle.Render("Нажмите Enter для перехода в меню или Ctrl+C для выхода")
}
