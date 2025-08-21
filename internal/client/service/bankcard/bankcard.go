package bankcard

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/mapper"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

// BankCardManagerIface описывает интерфейс управления банковскими картами.
type BankCardManagerIface interface {
	CreateBankCard(ctx context.Context, card *model.BankCard) error
	GetBankCardByID(ctx context.Context, id string) (*model.BankCard, error)
	GetBankCards(ctx context.Context) ([]model.BankCard, error)
	UpdateBankCard(ctx context.Context, card *model.BankCard) error
	DeleteBankCard(ctx context.Context, id string) error
	SetClient(client pb.BankCardServiceClient)
}

// BankCardManager управляет CRUD операциями с банковскими картами,
// взаимодействует с сервером по gRPC и логирует операции.
type BankCardManager struct {
	logger *zap.Logger
	client pb.BankCardServiceClient // для инъекции моков в тестах
}

// NewBankCardManager создаёт новый BankCardManager.
func NewBankCardManager(logger *zap.Logger) *BankCardManager {
	return &BankCardManager{
		logger: logger,
	}
}

// SetClient позволяет установить кастомный (например, моковый) gRPC-клиент.
func (m *BankCardManager) SetClient(client pb.BankCardServiceClient) {
	m.client = client
}

// CreateBankCard создаёт новую банковскую карту на сервере.
func (m *BankCardManager) CreateBankCard(ctx context.Context, card *model.BankCard) error {
	m.logger.Debug("CreateBankCard request started",
		zap.String("userID", card.UserID),
		zap.String("title", card.Title),
	)

	req := &pb.CreateBankCardRequest{}
	req.SetBankCard(mapper.BankCardToPB(card))

	_, err := m.client.CreateBankCard(ctx, req)
	if err != nil {
		m.logger.Error("CreateBankCard RPC failed", zap.Error(err))
		return fmt.Errorf("CreateBankCard RPC failed: %w", err)
	}

	m.logger.Info("CreateBankCard succeeded",
		zap.String("userID", card.UserID),
		zap.String("title", card.Title),
	)
	return nil
}

// GetBankCardByID получает данные банковской карты по ID.
func (m *BankCardManager) GetBankCardByID(ctx context.Context, id string) (*model.BankCard, error) {
	m.logger.Debug("GetBankCardByID request started",
		zap.String("bankCardID", id),
	)

	req := &pb.GetBankCardByIDRequest{}
	req.SetId(id)

	resp, err := m.client.GetBankCardByID(ctx, req)
	if err != nil {
		m.logger.Error("GetBankCardByID RPC failed", zap.Error(err))
		return nil, fmt.Errorf("GetBankCardByID RPC failed: %w", err)
	}

	card := mapper.BankCardFromPB(resp.GetBankCard())

	m.logger.Info("GetBankCardByID succeeded",
		zap.String("bankCardID", id),
	)
	return card, nil
}

// GetBankCards получает список банковских карт пользователя.
func (m *BankCardManager) GetBankCards(ctx context.Context) ([]model.BankCard, error) {
	m.logger.Debug("GetBankCards request started")

	resp, err := m.client.GetBankCards(ctx, &emptypb.Empty{})
	if err != nil {
		m.logger.Error("GetBankCards RPC failed", zap.Error(err))
		return nil, fmt.Errorf("GetBankCards RPC failed: %w", err)
	}

	cards := make([]model.BankCard, 0, len(resp.GetBankCards()))
	for _, pbCard := range resp.GetBankCards() {
		cards = append(cards, *mapper.BankCardFromPB(pbCard))
	}

	m.logger.Info("GetBankCards succeeded",
		zap.Int("count", len(cards)),
	)
	return cards, nil
}

// UpdateBankCard обновляет существующую банковскую карту на сервере.
func (m *BankCardManager) UpdateBankCard(ctx context.Context, card *model.BankCard) error {
	m.logger.Debug("UpdateBankCard request started",
		zap.String("bankCardID", card.ID),
	)

	req := &pb.UpdateBankCardRequest{}
	req.SetBankCard(mapper.BankCardToPB(card))

	_, err := m.client.UpdateBankCard(ctx, req)
	if err != nil {
		m.logger.Error("UpdateBankCard RPC failed", zap.Error(err))
		return fmt.Errorf("UpdateBankCard RPC failed: %w", err)
	}

	m.logger.Info("UpdateBankCard succeeded",
		zap.String("bankCardID", card.ID),
	)
	return nil
}

// DeleteBankCard удаляет банковскую карту по ID.
func (m *BankCardManager) DeleteBankCard(ctx context.Context, id string) error {
	m.logger.Debug("DeleteBankCard request started",
		zap.String("bankCardID", id),
	)

	req := &pb.DeleteBankCardRequest{}
	req.SetId(id)

	_, err := m.client.DeleteBankCard(ctx, req)
	if err != nil {
		m.logger.Error("DeleteBankCard RPC failed", zap.Error(err))
		return fmt.Errorf("DeleteBankCard RPC failed: %w", err)
	}

	m.logger.Info("DeleteBankCard succeeded",
		zap.String("bankCardID", id),
	)
	return nil
}
