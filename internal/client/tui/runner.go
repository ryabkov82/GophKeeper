package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

// Run запускает клиентское TUI-приложение
func Run(ctx context.Context, services *AppServices) error {

	model := NewModel(ctx, services)
	p := tea.NewProgram(model, tea.WithAltScreen())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		_, err := p.Run()
		done <- err
	}()

	select {
	case sig := <-sigCh:
		services.Logger.Warn("Получен сигнал завершения", zap.String("signal", sig.String()))

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownCh := make(chan struct{})
		go func() {
			_ = services.ConnManager.Close()
			p.Quit()
			close(shutdownCh)
		}()

		select {
		case <-shutdownCh:
			services.Logger.Info("Корректное завершение завершено")
		case <-shutdownCtx.Done():
			services.Logger.Warn("Превышено время ожидания завершения")
		}
		return nil

	case err := <-done:
		if err != nil {
			return fmt.Errorf("ошибка при выполнении TUI: %w", err)
		}
		services.Logger.Info("TUI завершился корректно")
		return nil
	}
}
