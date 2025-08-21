package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/tuiiface"
	"go.uber.org/zap"
)

// Реализация Program интерфейса просто через bubbletea.Program
type teaProgram struct {
	prog *tea.Program
}

// Run запускает выполнение программы bubbletea.
func (t *teaProgram) Run() (tea.Model, error) {
	return t.prog.Run()
}

// Quit корректно завершает выполнение программы bubbletea.
func (t *teaProgram) Quit() {
	t.prog.Quit()
}

// DefaultProgramFactory создаёт bubbletea-программу с нужными опциями
func DefaultProgramFactory(m tea.Model) tuiiface.Program {
	return &teaProgram{prog: tea.NewProgram(m, tea.WithAltScreen())}
}

// Run запускает клиентское TUI-приложение с переданными зависимостями.
//
// Аргументы:
//   - ctx: контекст для управления жизненным циклом приложения (может быть отменён из внешнего кода).
//   - services: контейнер зависимостей (логгер, менеджер соединения и др.).
//   - newProgram: фабрика для создания объекта Program на основе модели TUI.
//   - signals: канал для получения сигналов ОС (например, SIGINT, SIGTERM).
//
// Логика:
//   - Создаётся модель приложения с помощью NewModel.
//   - Создаётся и запускается программа TUI в отдельной горутине.
//   - В select ждёт либо сигнал завершения, либо завершение TUI.
//   - При получении сигнала:
//   - Выполняется graceful shutdown: закрывается соединение, вызывается Quit у TUI.
//   - Логируются события завершения.
//   - При завершении TUI с ошибкой возвращается обёрнутая ошибка.
//   - При успешном завершении возвращается nil.
func Run(
	ctx context.Context,
	services *app.AppServices,
	newProgram tuiiface.ProgramFactory,
) error {
	model := NewModel(ctx, ModelServices{
		Auth:       services,
		Credential: services,
		Bankcard:   services,
		TextData:   services,
		BinaryData: services,
	})
	p := newProgram(model)

	done := make(chan error, 1)
	go func() {
		_, err := p.Run()
		done <- err
	}()

	select {
	case <-ctx.Done():
		services.Logger.Warn("Контекст отменён, завершаем TUI")

		// Просто вызываем Close() - внутри него уже реализован таймаут и идемпотентность
		if err := services.Close(); err != nil {
			services.Logger.Error("Ошибка при закрытии ресурсов", zap.Error(err))
		}

		p.Quit()
		services.Logger.Info("Завершение после отмены контекста")

		return nil

	case err := <-done:
		if err != nil {
			return fmt.Errorf("ошибка при выполнении TUI: %w", err)
		}
		services.Logger.Info("TUI завершился корректно")
		return nil
	}
}
