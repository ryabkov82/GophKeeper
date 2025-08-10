package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/app"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Мок для auth.AuthManager
type mockAuthManager struct {
	registerErr error
	loginErr    error
}

func (m *mockAuthManager) Register(ctx context.Context, login, password string) error {
	return m.registerErr
}

func (m *mockAuthManager) Login(ctx context.Context, login, password string) error {
	return m.loginErr
}

// Помощник для создания модели с заполненными моками
func makeTestModel(t *testing.T, authMgr *mockAuthManager) Model {
	m := Model{
		ctx:      context.Background(),
		services: &app.AppServices{AuthManager: authMgr},
	}
	m = initRegisterForm(m)
	return m
}

func TestUpdateRegister_FocusSwitching(t *testing.T) {
	authMgr := &mockAuthManager{}
	m := makeTestModel(t, authMgr)

	tests := []struct {
		name          string
		keyMsg        tea.KeyMsg
		expectedFocus int
	}{
		{"Tab advances focus", tea.KeyMsg{Type: tea.KeyTab}, 1},
		{"ShiftTab moves focus back", tea.KeyMsg{Type: tea.KeyShiftTab}, 2},
		{"Down arrow advances focus", tea.KeyMsg{Type: tea.KeyDown}, 1},
		{"Up arrow moves focus back", tea.KeyMsg{Type: tea.KeyUp}, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m = initRegisterForm(m) // инициализируем inputs и сбрасываем состояние
			m.focusedInput = 0      // стартуем с 0
			m, _ = updateRegister(m, tc.keyMsg)
			assert.Equal(t, tc.expectedFocus, m.focusedInput)
		})
	}
}

func TestUpdateRegister_EnterPasswordsMismatch(t *testing.T) {
	authMgr := &mockAuthManager{}
	m := makeTestModel(t, authMgr)

	m.inputs[0].SetValue("user")
	m.inputs[1].SetValue("pass1")
	m.inputs[2].SetValue("pass2")

	m.focusedInput = 2

	m, cmd := updateRegister(m, tea.KeyMsg{Type: tea.KeyEnter})

	assert.Nil(t, cmd)
	require.NotNil(t, m.registerErr)
	assert.Equal(t, "пароли не совпадают", m.registerErr.Error())
}

func TestUpdateRegister_EnterPasswordsMatch(t *testing.T) {
	authMgr := &mockAuthManager{}
	m := makeTestModel(t, authMgr)

	tests := []struct {
		name         string
		registerErr  error
		expectType   interface{}
		expectErrStr string
	}{
		{"Success", nil, RegisterSuccessMsg{}, ""},
		{"Failure", errors.New("fail"), RegisterFailedMsg{}, "fail"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m.inputs[0].SetValue("user")
			m.inputs[1].SetValue("pass")
			m.inputs[2].SetValue("pass")
			m.focusedInput = 2

			m.services.AuthManager = authMgr
			authMgr.registerErr = tc.registerErr

			_, cmd := updateRegister(m, tea.KeyMsg{Type: tea.KeyEnter})
			require.NotNil(t, cmd)

			msg := cmd()

			switch v := msg.(type) {
			case tea.BatchMsg:
				var found bool
				for _, cmdFunc := range v {
					innerMsg := cmdFunc()
					switch inner := innerMsg.(type) {
					case RegisterSuccessMsg:
						found = true
						assert.IsType(t, tc.expectType, inner)
						assert.Empty(t, tc.expectErrStr)
					case RegisterFailedMsg:
						found = true
						assert.IsType(t, tc.expectType, inner)
						assert.EqualError(t, inner.Err, tc.expectErrStr)
					}
				}
				if !found {
					t.Fatalf("BatchMsg does not contain expected RegisterSuccessMsg or RegisterFailedMsg")
				}
			default:
				t.Fatalf("Unexpected message type: %T", v)
			}
		})
	}
}

func TestRenderRegister_ShowsError(t *testing.T) {
	authMgr := &mockAuthManager{}
	m := makeTestModel(t, authMgr)
	m.registerErr = errors.New("error message")

	output := renderRegister(m)
	assert.Contains(t, output, "error message")
}

func TestUpdateRegisterSuccess_KeyHandling(t *testing.T) {
	tests := []struct {
		name          string
		keyMsg        tea.KeyMsg
		expectedState string
		expectCmd     bool
	}{
		{"Enter changes state to menu", tea.KeyMsg{Type: tea.KeyEnter}, "menu", false},
		{"Ctrl+C quits", tea.KeyMsg{Type: tea.KeyCtrlC}, "registerSuccess", true},
		{"Other keys no change", tea.KeyMsg{Type: tea.KeyTab}, "registerSuccess", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := Model{currentState: "registerSuccess"}
			newM, cmd := updateRegisterSuccess(m, tc.keyMsg)
			assert.Equal(t, tc.expectedState, newM.currentState)
			if tc.expectCmd {
				assert.NotNil(t, cmd)
			} else {
				assert.Nil(t, cmd)
			}
		})
	}
}
