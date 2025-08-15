package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Мок BankCardService
type mockBankCardService struct {
	mock.Mock
}

func (m *mockBankCardService) Create(ctx context.Context, card *model.BankCard) error {
	return m.Called(ctx, card).Error(0)
}

func (m *mockBankCardService) GetByID(ctx context.Context, id string) (*model.BankCard, error) {
	args := m.Called(ctx, id)
	c := args.Get(0)
	if c == nil {
		return nil, args.Error(1)
	}
	return c.(*model.BankCard), args.Error(1)
}

func (m *mockBankCardService) GetByUserID(ctx context.Context, userID string) ([]model.BankCard, error) {
	args := m.Called(ctx, userID)
	cards := args.Get(0)
	if cards == nil {
		return nil, args.Error(1)
	}
	return cards.([]model.BankCard), args.Error(1)
}

func (m *mockBankCardService) Update(ctx context.Context, card *model.BankCard) error {
	return m.Called(ctx, card).Error(0)
}

func (m *mockBankCardService) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

// Мок jwtauth.FromContext
func mockJWTContext(userID string) context.Context {
	return jwtauth.WithUserID(context.Background(), userID)
}

func TestCreateBankCard_Success(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	ctx := mockJWTContext(userID)

	req := &pb.CreateBankCardRequest{}
	cardProto := &pb.BankCard{}
	cardProto.SetTitle("My Card")
	cardProto.SetCardholderName("John Doe")
	cardProto.SetCardNumber("1234123412341234")
	cardProto.SetExpiryDate("12/30")
	cardProto.SetCvv("123")
	cardProto.SetMetadata("{}")

	req.SetBankCard(cardProto)

	mockSvc.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := h.CreateBankCard(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, req.GetBankCard().GetTitle(), resp.GetBankCard().GetTitle())
}

func TestCreateBankCard_Unauthenticated(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	ctx := context.Background() // нет userID

	req := &pb.CreateBankCardRequest{}
	_, err := h.CreateBankCard(ctx, req)
	assert.Error(t, err)
}

func TestGetBankCardByID_NotFound(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	cardID := uuid.NewString()
	ctx := mockJWTContext(userID)

	mockSvc.On("GetByID", mock.Anything, cardID).Return(nil, nil)

	req := &pb.GetBankCardByIDRequest{}
	req.SetId(cardID)

	_, err := h.GetBankCardByID(ctx, req)
	assert.Error(t, err)
}

func TestGetBankCards_Success(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	ctx := mockJWTContext(userID)

	cards := []model.BankCard{
		{ID: uuid.NewString(), UserID: userID, Title: "Card1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.NewString(), UserID: userID, Title: "Card2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	mockSvc.On("GetByUserID", mock.Anything, userID).Return(cards, nil)

	resp, err := h.GetBankCards(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, resp.GetBankCards(), 2)
}

func TestUpdateBankCard_Success(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	cardID := uuid.NewString()
	ctx := mockJWTContext(userID)

	card := &model.BankCard{ID: cardID, UserID: userID}

	mockSvc.On("GetByID", mock.Anything, cardID).Return(card, nil)
	mockSvc.On("Update", mock.Anything, mock.Anything).Return(nil)

	req := &pb.UpdateBankCardRequest{}
	cardProto := &pb.BankCard{}
	cardProto.SetId(cardID)
	cardProto.SetTitle("Updated")
	req.SetBankCard(cardProto)

	resp, err := h.UpdateBankCard(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, cardID, resp.GetBankCard().GetId())
}

func TestDeleteBankCard_Success(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	cardID := uuid.NewString()
	ctx := mockJWTContext(userID)

	card := &model.BankCard{ID: cardID, UserID: userID}

	mockSvc.On("GetByID", mock.Anything, cardID).Return(card, nil)
	mockSvc.On("Delete", mock.Anything, cardID).Return(nil)

	reqDel := &pb.DeleteBankCardRequest{}
	reqDel.SetId(cardID)

	resp, err := h.DeleteBankCard(ctx, reqDel)
	assert.NoError(t, err)
	assert.True(t, resp.GetSuccess())
}

func TestCreateBankCard_ServiceError(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	ctx := mockJWTContext(userID)

	req := &pb.CreateBankCardRequest{}
	cardProto := &pb.BankCard{}
	cardProto.SetTitle("My Card")
	cardProto.SetCardholderName("John Doe")
	cardProto.SetCardNumber("1234123412341234")
	cardProto.SetExpiryDate("12/30")
	cardProto.SetCvv("123")
	cardProto.SetMetadata("{}")

	req.SetBankCard(cardProto)

	mockSvc.On("Create", mock.Anything, mock.Anything).Return(errors.New("service error"))

	_, err := h.CreateBankCard(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service error")
}

func TestGetBankCardByID_ServiceError(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	cardID := uuid.NewString()
	ctx := mockJWTContext(userID)

	mockSvc.On("GetByID", mock.Anything, cardID).Return(nil, errors.New("service error"))

	req := &pb.GetBankCardByIDRequest{}
	req.SetId(cardID)

	_, err := h.GetBankCardByID(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service error")
}

func TestUpdateBankCard_ServiceError(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	cardID := uuid.NewString()
	ctx := mockJWTContext(userID)

	existing := &model.BankCard{ID: cardID, UserID: userID}
	mockSvc.On("GetByID", mock.Anything, cardID).Return(existing, nil)
	mockSvc.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))

	req := &pb.UpdateBankCardRequest{}
	cardProto := &pb.BankCard{}
	cardProto.SetId(cardID)
	cardProto.SetTitle("Updated")
	req.SetBankCard(cardProto)

	_, err := h.UpdateBankCard(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update error")
}

func TestDeleteBankCard_ServiceError(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	cardID := uuid.NewString()
	ctx := mockJWTContext(userID)

	existing := &model.BankCard{ID: cardID, UserID: userID}
	mockSvc.On("GetByID", mock.Anything, cardID).Return(existing, nil)
	mockSvc.On("Delete", mock.Anything, cardID).Return(errors.New("delete error"))

	reqDel := &pb.DeleteBankCardRequest{}
	reqDel.SetId(cardID)

	_, err := h.DeleteBankCard(ctx, reqDel)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete error")
}

func TestGetBankCards_ServiceError(t *testing.T) {
	mockSvc := new(mockBankCardService)
	logger := zap.NewNop()
	h := handlers.NewBankCardHandler(mockSvc, logger)

	userID := uuid.NewString()
	ctx := mockJWTContext(userID)

	mockSvc.On("GetByUserID", mock.Anything, userID).Return(nil, errors.New("service error"))

	_, err := h.GetBankCards(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service error")
}
