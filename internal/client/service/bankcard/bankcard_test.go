package bankcard_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/client/service/bankcard"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// testHelper содержит общие зависимости для тестов
type testHelper struct {
	ctrl       *gomock.Controller
	mockClient *mocks.MockBankCardServiceClient
	manager    *bankcard.BankCardManager
}

// setupTest инициализирует общие зависимости для тестов
func setupTest(t *testing.T) *testHelper {
	ctrl := gomock.NewController(t)
	mockClient := mocks.NewMockBankCardServiceClient(ctrl)
	logger := zap.NewNop()
	manager := bankcard.NewBankCardManager(logger)
	manager.SetClient(mockClient)

	return &testHelper{
		ctrl:       ctrl,
		mockClient: mockClient,
		manager:    manager,
	}
}

// teardownTest освобождает ресурсы после теста
func (th *testHelper) teardownTest() {
	th.ctrl.Finish()
}

func TestBankCardManager_CreateBankCard(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	testCard := &model.BankCard{
		Title:          "Test Card",
		CardholderName: "TEST USER",
		CardNumber:     "4111111111111111",
		ExpiryDate:     "12/25",
		CVV:            "123",
	}

	t.Run("Success", func(t *testing.T) {
		// Мок просто возвращает успешный ответ без данных
		th.mockClient.EXPECT().
			CreateBankCard(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&pb.CreateBankCardResponse{}, nil)

		// Сохраняем исходное состояние (для проверки, что не изменилось)
		originalCard := *testCard

		err := th.manager.CreateBankCard(context.Background(), testCard)

		// Проверяем что:
		// 1. Нет ошибки
		// 2. Исходный объект не изменился
		assert.NoError(t, err)
		assert.Equal(t, originalCard.Title, testCard.Title)
		assert.Equal(t, originalCard.CardNumber, testCard.CardNumber)
		assert.Equal(t, originalCard.CVV, testCard.CVV)
		assert.Empty(t, testCard.ID)                // ID не должен устанавливаться клиентом
		assert.True(t, testCard.CreatedAt.IsZero()) // CreatedAt не должен устанавливаться клиентом
		assert.True(t, testCard.UpdatedAt.IsZero()) // UpdatedAt не должен устанавливаться клиентом
	})

	t.Run("Error from server", func(t *testing.T) {
		th.mockClient.EXPECT().
			CreateBankCard(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("server error"))

		err := th.manager.CreateBankCard(context.Background(), &model.BankCard{})
		assert.Error(t, err)
	})
}

func TestBankCardManager_GetBankCardByID(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	cardID := uuid.NewString()
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		// Настройка ожидаемого ответа
		resp := &pb.GetBankCardByIDResponse{}
		pbCard := &pb.BankCard{}
		pbCard.SetId(cardID)
		pbCard.SetTitle("Test Card")
		pbCard.SetCardholderName("TEST USER")
		pbCard.SetCardNumber("encrypted")
		pbCard.SetExpiryDate("12/25")
		pbCard.SetCvv("encrypted")
		pbCard.SetCreatedAt(timestamppb.New(now))
		pbCard.SetUpdatedAt(timestamppb.New(now))
		resp.SetBankCard(pbCard)

		// Настройка мока
		req := &pb.GetBankCardByIDRequest{}
		req.SetId(cardID)
		th.mockClient.EXPECT().
			GetBankCardByID(gomock.Any(), req).
			Return(resp, nil)

		card, err := th.manager.GetBankCardByID(context.Background(), cardID)
		assert.NoError(t, err)
		assert.Equal(t, cardID, card.ID)
		assert.Equal(t, "Test Card", card.Title)
	})

	t.Run("Not found", func(t *testing.T) {
		th.mockClient.EXPECT().
			GetBankCardByID(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("not found"))

		card, err := th.manager.GetBankCardByID(context.Background(), "invalid-id")
		assert.Error(t, err)
		assert.Nil(t, card)
	})
}

func TestBankCardManager_GetBankCards(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	now := time.Now()
	pbCard1 := &pb.BankCard{}
	pbCard1.SetId(uuid.NewString())
	pbCard1.SetTitle("Card 1")
	pbCard1.SetCreatedAt(timestamppb.New(now))

	pbCard2 := &pb.BankCard{}
	pbCard2.SetId(uuid.NewString())
	pbCard2.SetTitle("Card 2")
	pbCard2.SetCreatedAt(timestamppb.New(now.Add(-24 * time.Hour)))

	t.Run("Success with cards", func(t *testing.T) {
		resp := &pb.GetBankCardsResponse{}
		resp.SetBankCards([]*pb.BankCard{pbCard1, pbCard2})

		th.mockClient.EXPECT().
			GetBankCards(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
			Return(resp, nil)

		cards, err := th.manager.GetBankCards(context.Background())

		assert.NoError(t, err)
		assert.Len(t, cards, 2)
		assert.Equal(t, pbCard1.GetId(), cards[0].ID)
		assert.Equal(t, pbCard1.GetTitle(), cards[0].Title)
		assert.Equal(t, pbCard2.GetId(), cards[1].ID)
		assert.Equal(t, pbCard2.GetTitle(), cards[1].Title)
	})

	t.Run("Empty list", func(t *testing.T) {
		resp := &pb.GetBankCardsResponse{}
		resp.SetBankCards([]*pb.BankCard{})

		th.mockClient.EXPECT().
			GetBankCards(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
			Return(resp, nil)

		cards, err := th.manager.GetBankCards(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, cards)
	})

	t.Run("Error from server", func(t *testing.T) {
		th.mockClient.EXPECT().
			GetBankCards(gomock.Any(), &emptypb.Empty{}, gomock.Any()).
			Return(nil, errors.New("server error"))

		cards, err := th.manager.GetBankCards(context.Background())

		assert.Error(t, err)
		assert.Nil(t, cards)
	})
}

func TestBankCardManager_UpdateBankCard(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	testCard := &model.BankCard{
		ID:    uuid.NewString(),
		Title: "Updated Card",
	}

	t.Run("Success", func(t *testing.T) {
		th.mockClient.EXPECT().
			UpdateBankCard(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&pb.UpdateBankCardResponse{}, nil)

		originalCard := *testCard
		err := th.manager.UpdateBankCard(context.Background(), testCard)

		assert.NoError(t, err)
		// Проверяем, что объект не изменился
		assert.Equal(t, originalCard.ID, testCard.ID)
		assert.Equal(t, originalCard.Title, testCard.Title)
	})

	t.Run("Error from server", func(t *testing.T) {
		th.mockClient.EXPECT().
			UpdateBankCard(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("server error"))

		err := th.manager.UpdateBankCard(context.Background(), &model.BankCard{})

		assert.Error(t, err)
	})
}

func TestBankCardManager_DeleteBankCard(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	cardID := uuid.NewString()

	t.Run("Success", func(t *testing.T) {
		req := &pb.DeleteBankCardRequest{}
		req.SetId(cardID)
		th.mockClient.EXPECT().
			DeleteBankCard(gomock.Any(), req, gomock.Any()).
			Return(&pb.DeleteBankCardResponse{}, nil)

		err := th.manager.DeleteBankCard(context.Background(), cardID)

		assert.NoError(t, err)
	})

	t.Run("Error from server", func(t *testing.T) {
		th.mockClient.EXPECT().
			DeleteBankCard(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("server error"))

		err := th.manager.DeleteBankCard(context.Background(), cardID)

		assert.Error(t, err)
	})
}
