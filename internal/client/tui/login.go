package tui

import (
	"context"
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/auth"
)

var loginFieldLabels = []string{
	"Логин",
	"Пароль",
}

func initLoginForm(m Model) Model {
	m.currentState = "login"
	m.inputs = make([]textinput.Model, 2)

	for i := range m.inputs {
		m.inputs[i] = newInputField("")
	}
	m.inputs[0].Focus()
	m.inputs[1].EchoMode = textinput.EchoPassword

	m.focusedInput = 0
	m.loginErr = nil

	return m
}

func updateLogin(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.focusedInput == len(m.inputs)-1 {
				// Логика авторизации
				login := m.inputs[0].Value()
				password := m.inputs[1].Value()

				if strings.TrimSpace(login) == "" || strings.TrimSpace(password) == "" {
					m.loginErr = errors.New("логин и пароль не должны быть пустыми")
					return m, nil
				}

				return m, tea.Batch(
					tea.Printf("Авторизация..."),
					loginUser(m.ctx, m.services.AuthManager, login, password),
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

	case LoginSuccessMsg:
		// Вместо перехода сразу в меню — переключаем состояние на loginSuccess
		m.currentState = "loginSuccess"
		m.loginErr = nil
		return m, nil

	case LoginFailedMsg:
		m.loginErr = msg.Err
		return m, nil
	}

	var cmd tea.Cmd
	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
	return m, cmd
}

func renderLogin(m Model) string {
	var builder strings.Builder

	builder.WriteString(titleStyle.Render("Авторизация"))
	builder.WriteString("\n\n")

	for i, input := range m.inputs {
		label := loginFieldLabels[i] + ": "
		if i == m.focusedInput {
			label = activeFieldStyle.Render(label)
		} else {
			label = inactiveFieldStyle.Render(label)
		}

		builder.WriteString(label + input.View() + "\n")
	}

	if m.loginErr != nil {
		builder.WriteString("\n" + errorStyle.Render("Ошибка: "+m.loginErr.Error()))
	}

	builder.WriteString("\n" + hintStyle.Render(
		"Tab: переключение • Enter: подтвердить • Esc: назад • Ctrl+C: выход",
	))

	return builder.String()
}

func updateLoginSuccess(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.currentState = "menu"
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func renderLoginSuccess(m Model) string {
	return titleStyle.Render("Авторизация") + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("Успешный вход!") + "\n\n" +
		hintStyle.Render("Нажмите Enter для перехода в меню или Ctrl+C для выхода")
}

// Команда для авторизации
func loginUser(ctx context.Context, authManager auth.AuthManagerIface, login, password string) tea.Cmd {
	return func() tea.Msg {
		err := authManager.Login(ctx, login, password)
		if err != nil {
			return LoginFailedMsg{Err: err}
		}
		return LoginSuccessMsg{}
	}
}

// Сообщения авторизации
type LoginSuccessMsg struct{}
type LoginFailedMsg struct {
	Err error
}
