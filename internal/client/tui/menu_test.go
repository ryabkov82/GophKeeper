package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Вспомогательная функция для создания модели с дефолтным меню
func makeTestMenuModel() Model {
	m := Model{
		menuItems: []menuItem{
			{title: "Login", description: "Вход в систему"},
			{title: "Register", description: "Регистрация"},
			{title: "View Data", description: "Просмотр данных"},
			{title: "Exit", description: "Выход из программы"},
		},
		menuCursor:   0,
		currentState: "menu",
	}
	return m
}

func TestUpdateMenu_KeyHandling(t *testing.T) {
	baseModel := func() Model {
		return Model{
			menuItems: []menuItem{
				{title: "Login", description: "Вход"},
				{title: "Register", description: "Регистрация"},
				{title: "View Data", description: "Просмотр"},
				{title: "Exit", description: "Выход"},
			},
			menuCursor:   0,
			currentState: "menu",
		}
	}

	tests := []struct {
		name           string
		initialCursor  int
		keyMsg         tea.KeyMsg
		expectedCursor int
		expectedState  string
		expectQuit     bool
	}{
		{
			name:           "Arrow down moves cursor down",
			initialCursor:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 1,
			expectedState:  "menu",
		},
		{
			name:           "Arrow up moves cursor up",
			initialCursor:  2,
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 1,
			expectedState:  "menu",
		},
		{
			name:           "Cursor does not go below 0",
			initialCursor:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyUp},
			expectedCursor: 0,
			expectedState:  "menu",
		},
		{
			name:           "Cursor does not go above max",
			initialCursor:  3,
			keyMsg:         tea.KeyMsg{Type: tea.KeyDown},
			expectedCursor: 3,
			expectedState:  "menu",
		},
		{
			name:           "Enter on Login sets state login",
			initialCursor:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyEnter},
			expectedCursor: 0,
			expectedState:  "login",
		},
		{
			name:           "Enter on Register sets state register",
			initialCursor:  1,
			keyMsg:         tea.KeyMsg{Type: tea.KeyEnter},
			expectedCursor: 1,
			expectedState:  "register",
		},
		{
			name:           "Enter on View Data sets state view_data",
			initialCursor:  2,
			keyMsg:         tea.KeyMsg{Type: tea.KeyEnter},
			expectedCursor: 2,
			expectedState:  "view_data",
		},
		{
			name:           "Enter on Exit returns quit command",
			initialCursor:  3,
			keyMsg:         tea.KeyMsg{Type: tea.KeyEnter},
			expectedCursor: 3,
			expectedState:  "menu",
			expectQuit:     true,
		},
		{
			name:           "Ctrl+C returns quit command",
			initialCursor:  0,
			keyMsg:         tea.KeyMsg{Type: tea.KeyCtrlC},
			expectedCursor: 0,
			expectedState:  "menu",
			expectQuit:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := baseModel()
			m.menuCursor = tc.initialCursor

			m, cmd := updateMenu(m, tc.keyMsg)

			assert.Equal(t, tc.expectedCursor, m.menuCursor)
			assert.Equal(t, tc.expectedState, m.currentState)

			if tc.expectQuit {
				require.NotNil(t, cmd)
				msg := cmd()
				_, ok := msg.(tea.QuitMsg)
				assert.True(t, ok)
			} else {
				assert.Nil(t, cmd)
			}
		})
	}
}

func TestUpdateMenu_EnterSelection(t *testing.T) {
	m := makeTestMenuModel()

	// выбрать Login — должен смениться currentState на "login"
	m.menuCursor = 0
	m, cmd := updateMenu(m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "login", m.currentState)
	assert.Nil(t, cmd)

	// выбрать Register — смена currentState на "register"
	m.menuCursor = 1
	m, cmd = updateMenu(m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "register", m.currentState)
	assert.Nil(t, cmd)

	// выбрать View Data — смена currentState на "view_data"
	m.menuCursor = 2
	m, cmd = updateMenu(m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "view_data", m.currentState)
	assert.Nil(t, cmd)

	// выбрать Exit — должна вернуться команда Quit
	m.menuCursor = 3
	m, cmd = updateMenu(m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok)
}

func TestUpdateMenu_CtrlC_Quits(t *testing.T) {
	m := makeTestMenuModel()
	_, cmd := updateMenu(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok)
}

func TestRenderMenu_Output(t *testing.T) {
	m := makeTestMenuModel()
	output := renderMenu(m)

	// Проверяем, что заголовок есть в выводе
	assert.Contains(t, output, "GophKeeper - Менеджер паролей")

	// Проверяем, что меню содержит все пункты
	for _, item := range m.menuItems {
		assert.Contains(t, output, item.title)
		assert.Contains(t, output, item.description)
	}

	// Проверяем, что подсказки есть в выводе
	assert.Contains(t, output, "↑/↓: навигация • Enter: выбор • Ctrl+C: выход")
}
