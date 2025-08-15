package handlers

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BankCardHandler реализует gRPC сервер для BankCardService
type BankCardHandler struct {
	pb.UnimplementedBankCardServiceServer
	service service.BankCardService
	logger  *zap.Logger
}

// NewBankCardHandler создает новый BankCardHandler с внедрением сервиса и логгера.
func NewBankCardHandler(srv service.BankCardService, logger *zap.Logger) *BankCardHandler {
	return &BankCardHandler{
		service: srv,
		logger:  logger,
	}
}

// CreateBankCard создает новую запись банковской карты.
func (h *BankCardHandler) CreateBankCard(ctx context.Context, req *pb.CreateBankCardRequest) (*pb.CreateBankCardResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("CreateBankCard request received",
		zap.String("userID", userID),
		zap.String("title", req.GetBankCard().GetTitle()),
	)

	card := &model.BankCard{
		UserID:         userID,
		Title:          req.GetBankCard().GetTitle(),
		CardholderName: req.GetBankCard().GetCardholderName(),
		CardNumber:     req.GetBankCard().GetCardNumber(),
		ExpiryDate:     req.GetBankCard().GetExpiryDate(),
		CVV:            req.GetBankCard().GetCvv(),
		Metadata:       req.GetBankCard().GetMetadata(),
	}

	err = h.service.Create(ctx, card)
	if err != nil {
		h.logger.Warn("CreateBankCard failed",
			zap.String("userID", userID),
			zap.String("title", req.GetBankCard().GetTitle()),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("CreateBankCard succeeded",
		zap.String("userID", userID),
		zap.String("bankCardID", card.ID),
	)

	resp := &pb.CreateBankCardResponse{}
	resp.SetBankCard(toProtoBankCard(card))
	return resp, nil

}

// GetBankCardByID возвращает данные банковской карты по ID.
func (h *BankCardHandler) GetBankCardByID(ctx context.Context, req *pb.GetBankCardByIDRequest) (*pb.GetBankCardByIDResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("GetBankCardByID request received",
		zap.String("userID", userID),
		zap.String("bankCardID", req.GetId()),
	)

	card, err := h.service.GetByID(ctx, req.GetId())
	if err != nil {
		h.logger.Warn("GetBankCardByID failed",
			zap.String("userID", userID),
			zap.String("bankCardID", req.GetId()),
			zap.Error(err),
		)
		return nil, err
	}
	if card == nil || card.UserID != userID {
		h.logger.Info("BankCard not found",
			zap.String("userID", userID),
			zap.String("bankCardID", req.GetId()),
		)
		return nil, status.Error(codes.NotFound, "bank card not found")
	}

	h.logger.Info("GetBankCardByID succeeded",
		zap.String("userID", userID),
		zap.String("bankCardID", req.GetId()),
	)

	resp := &pb.GetBankCardByIDResponse{}
	resp.SetBankCard(toProtoBankCard(card))
	return resp, nil
}

// GetBankCards возвращает все банковские карты пользователя.
func (h *BankCardHandler) GetBankCards(ctx context.Context, _ *emptypb.Empty) (*pb.GetBankCardsResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("GetBankCards request received",
		zap.String("userID", userID),
	)

	cards, err := h.service.GetByUserID(ctx, userID)
	if err != nil {
		h.logger.Warn("GetBankCards failed",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("GetBankCards succeeded",
		zap.String("userID", userID),
		zap.Int("count", len(cards)),
	)

	resp := &pb.GetBankCardsResponse{}
	for i := range cards {
		resp.SetBankCards(append(resp.GetBankCards(), toProtoBankCard(&cards[i])))
	}

	return resp, nil
}

// UpdateBankCard обновляет существующую запись банковской карты.
func (h *BankCardHandler) UpdateBankCard(ctx context.Context, req *pb.UpdateBankCardRequest) (*pb.UpdateBankCardResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("UpdateBankCard request received",
		zap.String("userID", userID),
		zap.String("bankCardID", req.GetBankCard().GetId()),
	)

	cardProto := req.GetBankCard()
	card := &model.BankCard{
		ID:             cardProto.GetId(),
		UserID:         userID,
		Title:          cardProto.GetTitle(),
		CardholderName: cardProto.GetCardholderName(),
		CardNumber:     cardProto.GetCardNumber(),
		ExpiryDate:     cardProto.GetExpiryDate(),
		CVV:            cardProto.GetCvv(),
		Metadata:       cardProto.GetMetadata(),
	}

	existing, err := h.service.GetByID(ctx, card.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil || existing.UserID != userID {
		return nil, status.Error(codes.NotFound, "bank card not found")
	}

	err = h.service.Update(ctx, card)
	if err != nil {
		h.logger.Warn("UpdateBankCard failed",
			zap.String("userID", userID),
			zap.String("bankCardID", cardProto.GetId()),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("UpdateBankCard succeeded",
		zap.String("userID", userID),
		zap.String("bankCardID", cardProto.GetId()),
	)

	resp := &pb.UpdateBankCardResponse{}
	resp.SetBankCard(toProtoBankCard(card))
	return resp, nil
}

// DeleteBankCard удаляет запись банковской карты по идентификатору.
func (h *BankCardHandler) DeleteBankCard(ctx context.Context, req *pb.DeleteBankCardRequest) (*pb.DeleteBankCardResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("DeleteBankCard request received",
		zap.String("userID", userID),
		zap.String("bankCardID", req.GetId()),
	)

	existing, err := h.service.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if existing == nil || existing.UserID != userID {
		return nil, status.Error(codes.NotFound, "bank card not found")
	}

	err = h.service.Delete(ctx, req.GetId())
	if err != nil {
		h.logger.Warn("DeleteBankCard failed",
			zap.String("userID", userID),
			zap.String("bankCardID", req.GetId()),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("DeleteBankCard succeeded",
		zap.String("userID", userID),
		zap.String("bankCardID", req.GetId()),
	)

	resp := &pb.DeleteBankCardResponse{}
	resp.SetSuccess(true)
	return resp, nil
}

// toProtoBankCard конвертирует модель BankCard в protobuf структуру BankCard.
func toProtoBankCard(c *model.BankCard) *pb.BankCard {
	card := &pb.BankCard{}
	card.SetId(c.ID)
	card.SetUserId(c.UserID)
	card.SetTitle(c.Title)
	card.SetCardholderName(c.CardholderName)
	card.SetCardNumber(c.CardNumber)
	card.SetExpiryDate(c.ExpiryDate)
	card.SetCvv(c.CVV)
	card.SetMetadata(c.Metadata)
	card.SetCreatedAt(timestamppb.New(c.CreatedAt))
	card.SetUpdatedAt(timestamppb.New(c.UpdatedAt))
	return card
}
