package grpc

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	api "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	api.UnimplementedAuthServiceServer
	service service.AuthService
}

// NewAuthHandler создаёт gRPC-хендлер с внедрённым AuthService
func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: authSvc,
	}
}

// Register реализует метод регистрации пользователя
func (h *AuthHandler) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	login := req.GetLogin()
	password := req.GetPassword()

	if err := h.service.Register(ctx, login, password); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "registration failed: %v", err)
	}

	resp := api.RegisterResponse{}
	resp.SetMessage("user registered successfully")
	return &resp, nil
}

// Login реализует метод входа пользователя
func (h *AuthHandler) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	login := req.GetLogin()
	password := req.GetPassword()

	token, err := h.service.Login(ctx, login, password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	resp := api.LoginResponse{}
	resp.SetAccessToken(token)
	return &resp, nil
}
