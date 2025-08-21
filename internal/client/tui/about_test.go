package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestAboutModel() Model {
	return Model{
		currentState: "about",
	}
}

func TestUpdateAbout_KeyHandling(t *testing.T) {
	tests := []struct {
		name          string
		keyMsg        tea.KeyMsg
		expectedState string
		expectQuit    bool
	}{
		{
			name:          "Enter returns to menu",
			keyMsg:        tea.KeyMsg{Type: tea.KeyEnter},
			expectedState: "menu",
		},
		{
			name:          "Esc returns to menu",
			keyMsg:        tea.KeyMsg{Type: tea.KeyEscape},
			expectedState: "menu",
		},
		{
			name:          "Ctrl+C quits",
			keyMsg:        tea.KeyMsg{Type: tea.KeyCtrlC},
			expectedState: "about",
			expectQuit:    true,
		},
		{
			name:          "Other keys have no effect",
			keyMsg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}},
			expectedState: "about",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := makeTestAboutModel()
			m, cmd := updateAbout(m, tc.keyMsg)

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

func TestRenderAbout_Output(t *testing.T) {
	origVersion, origDate := buildVersion, buildDate
	defer func() {
		buildVersion, buildDate = origVersion, origDate
	}()

	buildVersion = "1.0.0"
	buildDate = "2024-01-02"

	output := renderAbout(Model{})

	assert.Contains(t, output, "О программе")
	assert.Contains(t, output, "Версия: 1.0.0")
	assert.Contains(t, output, "Дата сборки: 2024-01-02")
	assert.Contains(t, output, "Enter/Esc: назад • Ctrl+C: выход")
}

func TestRenderAbout_EmptyValues(t *testing.T) {
	origVersion, origDate := buildVersion, buildDate
	defer func() {
		buildVersion, buildDate = origVersion, origDate
	}()

	buildVersion = ""
	buildDate = ""

	output := renderAbout(Model{})

	assert.Contains(t, output, "Версия: N/A")
	assert.Contains(t, output, "Дата сборки: N/A")
}
