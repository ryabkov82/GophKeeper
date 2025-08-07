package tui

import (
	"github.com/ryabkov82/gophkeeper/internal/client/auth"
	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"go.uber.org/zap"
)

// AppServices представляет контейнер всех зависимостей, необходимых для работы клиента.
// Это упрощает передачу сервисов в TUI и другие слои приложения, делает код более модульным
// и облегчает тестирование.
//
// Поля:
//   - AuthManager: отвечает за регистрацию, вход и управление токеном доступа пользователя.
//   - ConnManager: управляет подключением к gRPC-серверу.
//   - Logger: структурированный логгер для записи отладочной и диагностической информации.
type AppServices struct {
	AuthManager *auth.AuthManager
	ConnManager *connection.Manager
	Logger      *zap.Logger
	// Будущие зависимости:
	// DataService *data.Service
}
