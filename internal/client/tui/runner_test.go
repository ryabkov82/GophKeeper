package tui_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"github.com/ryabkov82/gophkeeper/internal/client/tui"
	"github.com/ryabkov82/gophkeeper/internal/client/tuiiface"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Мок реализации tuiiface.Program с управляемым поведением Run()
type mockProgram struct {
	runErr     error
	quitCalled bool

	// Если true, Run блокируется до вызова Quit
	runBlocks  bool
	runStarted chan struct{}
	quitCh     chan struct{}
}

func (m *mockProgram) Run() (tea.Model, error) {
	if m.runStarted != nil {
		close(m.runStarted) // сигнал что Run начался
	}

	if m.runBlocks {
		<-m.quitCh
		return nil, nil
	}
	return nil, m.runErr
}

func (m *mockProgram) Quit() {
	m.quitCalled = true
	if m.quitCh != nil {
		close(m.quitCh)
	}
}

// Мок ConnManager с проверкой вызова Close()
type mockConnManager struct {
	closeCalled bool
}

func (m *mockConnManager) Connect(ctx context.Context) (connection.GrpcConn, error) {
	return nil, nil
}

func (m *mockConnManager) Close() error {
	m.closeCalled = true
	return nil
}

func TestRun_Success(t *testing.T) {
	ctx := context.Background()

	services := &app.AppServices{
		Logger:      zap.NewNop(),
		ConnManager: &mockConnManager{},
	}

	mockProg := &mockProgram{
		runErr:    nil,
		runBlocks: false,
	}

	runFunc := func(m tea.Model) tuiiface.Program {
		return mockProg
	}

	err := tui.Run(ctx, services, runFunc)
	assert.NoError(t, err)
	assert.False(t, mockProg.quitCalled, "Quit не должен быть вызван при нормальном завершении")
	assert.False(t, services.ConnManager.(*mockConnManager).closeCalled, "Close не должен быть вызван без отмены контекста")
}

func TestRun_CtxCancel_Shutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mockConn := &mockConnManager{}

	services := &app.AppServices{
		Logger:      zap.NewNop(),
		ConnManager: mockConn,
	}

	mockProg := &mockProgram{
		runBlocks:  true,
		runStarted: make(chan struct{}),
		quitCh:     make(chan struct{}),
	}

	runFunc := func(m tea.Model) tuiiface.Program {
		return mockProg
	}

	errCh := make(chan error)
	go func() {
		errCh <- tui.Run(ctx, services, runFunc)
	}()

	<-mockProg.runStarted

	cancel()

	err := <-errCh
	assert.NoError(t, err)
	assert.True(t, mockProg.quitCalled, "Ожидается вызов Quit()")
	assert.True(t, mockConn.closeCalled, "Ожидается вызов Close() ConnManager")
}

func TestRun_RunError(t *testing.T) {
	ctx := context.Background()

	services := &app.AppServices{
		Logger:      zap.NewNop(),
		ConnManager: &mockConnManager{},
	}

	mockProg := &mockProgram{runErr: assert.AnError}

	runFunc := func(m tea.Model) tuiiface.Program {
		return mockProg
	}

	err := tui.Run(ctx, services, runFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), assert.AnError.Error())
	assert.False(t, mockProg.quitCalled, "Quit не должен быть вызван при ошибке Run")
	assert.False(t, services.ConnManager.(*mockConnManager).closeCalled, "Close не должен быть вызван без отмены контекста")
}
