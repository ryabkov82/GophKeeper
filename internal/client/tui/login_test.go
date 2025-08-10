package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestLoginModel(t *testing.T, authMgr *mockAuthManager) Model {
	m := Model{
		ctx:      context.Background(),
		services: &app.AppServices{AuthManager: authMgr},
	}
	m = initLoginForm(m)
	return m
}

func TestUpdateLogin_FocusSwitching(t *testing.T) {
	authMgr := &mockAuthManager{}
	m := makeTestLoginModel(t, authMgr)

	tests := []struct {
		name          string
		keyMsg        tea.KeyMsg
		expectedFocus int
	}{
		{"Tab advances focus", tea.KeyMsg{Type: tea.KeyTab}, 1},
		{"ShiftTab moves focus back", tea.KeyMsg{Type: tea.KeyShiftTab}, 1},
		{"Down arrow advances focus", tea.KeyMsg{Type: tea.KeyDown}, 1},
		{"Up arrow moves focus back", tea.KeyMsg{Type: tea.KeyUp}, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m.focusedInput = 0
			m, _ = updateLogin(m, tc.keyMsg)
			assert.Equal(t, tc.expectedFocus, m.focusedInput)
		})
	}
}

func TestUpdateLogin_EnterEmptyFields(t *testing.T) {
	authMgr := &mockAuthManager{}
	m := makeTestLoginModel(t, authMgr)

	// Пустой логин и пароль
	m.inputs[0].SetValue("")
	m.inputs[1].SetValue("")
	m.focusedInput = 1

	m, cmd := updateLogin(m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
	require.NotNil(t, m.loginErr)
	assert.Equal(t, "логин и пароль не должны быть пустыми", m.loginErr.Error())
}

func TestUpdateLogin_EnterValidCredentials(t *testing.T) {
	authMgr := &mockAuthManager{}
	m := makeTestLoginModel(t, authMgr)

	m.inputs[0].SetValue("user")
	m.inputs[1].SetValue("pass")
	m.focusedInput = 1
	m.services.AuthManager = authMgr

	tests := []struct {
		name         string
		loginErr     error
		expectType   interface{}
		expectErrStr string
	}{
		{"Success", nil, LoginSuccessMsg{}, ""},
		{"Failure", errors.New("fail"), LoginFailedMsg{}, "fail"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authMgr.loginErr = tc.loginErr

			_, cmd := updateLogin(m, tea.KeyMsg{Type: tea.KeyEnter})
			require.NotNil(t, cmd)

			msg := cmd()

			switch v := msg.(type) {
			case tea.BatchMsg:
				var found bool
				for _, cmdFunc := range v {
					innerMsg := cmdFunc()
					switch inner := innerMsg.(type) {
					case LoginSuccessMsg:
						found = true
						assert.IsType(t, tc.expectType, inner)
						assert.Empty(t, tc.expectErrStr)
					case LoginFailedMsg:
						found = true
						assert.IsType(t, tc.expectType, inner)
						assert.EqualError(t, inner.Err, tc.expectErrStr)
					}
				}
				if !found {
					t.Fatalf("BatchMsg does not contain expected LoginSuccessMsg or LoginFailedMsg")
				}
			default:
				t.Fatalf("Unexpected message type: %T", v)
			}
		})
	}
}

func TestRenderLogin_ShowsError(t *testing.T) {
	authMgr := &mockAuthManager{}
	m := makeTestLoginModel(t, authMgr)
	m.loginErr = errors.New("some error")

	output := renderLogin(m)
	assert.Contains(t, output, "some error")
}

func TestUpdateLoginSuccess_EnterChangesState(t *testing.T) {
	m := Model{currentState: "loginSuccess"}

	m, cmd := updateLoginSuccess(m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "menu", m.currentState)
	assert.Nil(t, cmd)
}

func TestUpdateLoginSuccess_CtrlCQuits(t *testing.T) {
	m := Model{currentState: "loginSuccess"}

	_, cmd := updateLoginSuccess(m, tea.KeyMsg{Type: tea.KeyCtrlC})

	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok, "expected tea.QuitMsg")

}
