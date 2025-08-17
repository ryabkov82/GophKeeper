package textdata_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/client/service/textdata"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// helper struct для тестов
type testHelper struct {
	ctrl       *gomock.Controller
	mockClient *mocks.MockTextDataServiceClient
	manager    *textdata.TextDataManager
}

func setupTest(t *testing.T) *testHelper {
	ctrl := gomock.NewController(t)
	mockClient := mocks.NewMockTextDataServiceClient(ctrl)
	logger := zap.NewNop()
	manager := textdata.NewTextDataManager(logger)
	manager.SetClient(mockClient)
	return &testHelper{
		ctrl:       ctrl,
		mockClient: mockClient,
		manager:    manager,
	}
}

func (th *testHelper) teardownTest() {
	th.ctrl.Finish()
}

func TestTextDataManager_CreateTextData(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	td := &model.TextData{
		Title:   "Note1",
		Content: []byte("secret"),
	}

	t.Run("Success", func(t *testing.T) {
		th.mockClient.EXPECT().
			CreateTextData(gomock.Any(), gomock.Any()).
			Return(&pb.CreateTextDataResponse{}, nil)

		err := th.manager.CreateTextData(context.Background(), td)
		assert.NoError(t, err)
	})

	t.Run("Too large content", func(t *testing.T) {
		tdLarge := &model.TextData{
			Title:   "Note2",
			Content: make([]byte, textdata.MaxContentSize+1),
		}
		err := th.manager.CreateTextData(context.Background(), tdLarge)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content too large")
	})

	t.Run("Server error", func(t *testing.T) {
		th.mockClient.EXPECT().
			CreateTextData(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("server error"))

		err := th.manager.CreateTextData(context.Background(), td)
		assert.Error(t, err)
	})
}

func TestTextDataManager_GetTextDataByID(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	id := uuid.NewString()
	content := []byte("secret content")

	t.Run("Success", func(t *testing.T) {
		pbtd := &pb.TextData{}
		pbtd.SetId(id)
		pbtd.SetTitle("Note1")
		pbtd.SetContent(content)

		resp := &pb.GetTextDataByIDResponse{}
		resp.SetTextData(pbtd)

		req := &pb.GetTextDataByIDRequest{}
		req.SetId(id)

		th.mockClient.EXPECT().
			GetTextDataByID(gomock.Any(), req).
			Return(resp, nil)

		td, err := th.manager.GetTextDataByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, id, td.ID)
		assert.Equal(t, content, td.Content)
	})
}

func TestTextDataManager_GetTextDataTitles(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	t.Run("Success", func(t *testing.T) {
		pbtd1 := &pb.TextData{}
		pbtd1.SetId("1")
		pbtd1.SetTitle("Note1")
		pbtd2 := &pb.TextData{}
		pbtd2.SetId("2")
		pbtd2.SetTitle("Note2")

		resp := &pb.GetTextDataTitlesResponse{}
		resp.SetTextDataTitles([]*pb.TextData{pbtd1, pbtd2})

		th.mockClient.EXPECT().
			GetTextDataTitles(gomock.Any(), gomock.Any()).
			Return(resp, nil)

		result, err := th.manager.GetTextDataTitles(context.Background())
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "1", result[0].ID)
		assert.Equal(t, "Note2", result[1].Title)
	})
}

func TestTextDataManager_UpdateTextData(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	td := &model.TextData{
		ID:      uuid.NewString(),
		Title:   "Updated",
		Content: []byte("updated content"),
	}

	t.Run("Success", func(t *testing.T) {
		th.mockClient.EXPECT().
			UpdateTextData(gomock.Any(), gomock.Any()).
			Return(&pb.UpdateTextDataResponse{}, nil)

		err := th.manager.UpdateTextData(context.Background(), td)
		assert.NoError(t, err)
	})

	t.Run("Too large content", func(t *testing.T) {
		td.Content = make([]byte, textdata.MaxContentSize+1)
		err := th.manager.UpdateTextData(context.Background(), td)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content too large")
	})
}

func TestTextDataManager_DeleteTextData(t *testing.T) {
	th := setupTest(t)
	defer th.teardownTest()

	id := uuid.NewString()

	t.Run("Success", func(t *testing.T) {
		req := &pb.DeleteTextDataRequest{}
		req.SetId(id)

		th.mockClient.EXPECT().
			DeleteTextData(gomock.Any(), req).
			Return(&pb.DeleteTextDataResponse{}, nil)

		err := th.manager.DeleteTextData(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("Server error", func(t *testing.T) {
		th.mockClient.EXPECT().
			DeleteTextData(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("server error"))

		err := th.manager.DeleteTextData(context.Background(), id)
		assert.Error(t, err)
	})
}
