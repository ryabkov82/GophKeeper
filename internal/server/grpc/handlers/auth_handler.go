package handlers

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	api "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler реализует gRPC-сервер для AuthService.
//
// Отвечает за обработку входящих gRPC-запросов, связанных с аутентификацией,
// и делегирует бизнес-логику соответствующему сервису.
//
// Поля:
//   - api.UnimplementedAuthServiceServer: встраиваемая заготовка gRPC-сервера,
//     необходима для совместимости с сгенерированным интерфейсом;
//   - service: интерфейс сервиса аутентификации, реализующий основную логику;
//   - Logger: zap-логгер для записи ошибок и событий.
type AuthHandler struct {
	api.UnimplementedAuthServiceServer
	service service.AuthService
	Logger  *zap.Logger
}

// NewAuthHandler создаёт gRPC-хендлер с внедрённым AuthService
func NewAuthHandler(authSvc service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service: authSvc,
		Logger:  log,
	}
}

// Register реализует метод регистрации пользователя
func (h *AuthHandler) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	login := req.GetLogin()
	password := req.GetPassword()

	h.Logger.Debug("Register request received",
		zap.String("login", login),
	)

	if err := h.service.Register(ctx, login, password); err != nil {
		h.Logger.Warn("User registration failed",
			zap.String("login", login),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.InvalidArgument, "registration failed: %v", err)
	}

	h.Logger.Info("User registered successfully",
		zap.String("login", login),
	)

	resp := api.RegisterResponse{}
	resp.SetMessage("user registered successfully")
	return &resp, nil
}

// Login реализует метод входа пользователя
func (h *AuthHandler) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	login := req.GetLogin()
	password := req.GetPassword()

	h.Logger.Debug("Login request received",
		zap.String("login", login),
	)

	token, err := h.service.Login(ctx, login, password)
	if err != nil {
		h.Logger.Warn("Login failed",
			zap.String("login", login),
			zap.Error(err),
		)
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	h.Logger.Info("User logged in successfully",
		zap.String("login", login),
	)

	resp := api.LoginResponse{}
	resp.SetAccessToken(token)
	return &resp, nil
}
