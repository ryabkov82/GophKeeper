package credential_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ryabkov82/gophkeeper/internal/client/service/credential"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto/mocks"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Создадим фиктивные данные для тестов
var testCredential = &model.Credential{
	ID:        "cred-1",
	UserID:    "user-1",
	Title:     "Test Title",
	Login:     "login@example.com",
	Password:  "encrypted-password",
	Metadata:  "metadata",
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

func setup(t *testing.T) (*credential.CredentialManager, *gomock.Controller, *mocks.MockCredentialServiceClient) {
	ctrl := gomock.NewController(t)
	mockClient := mocks.NewMockCredentialServiceClient(ctrl)

	logger := zaptest.NewLogger(t)

	manager := credential.NewCredentialManager(logger)

	// Внедрим мок-клиент (чтобы не создавать соединение в тестах)
	manager.SetClient(mockClient)

	return manager, ctrl, mockClient
}

func TestCreateCredential_Success(t *testing.T) {
	manager, ctrl, mockClient := setup(t)
	defer ctrl.Finish()

	mockClient.EXPECT().
		CreateCredential(gomock.Any(), gomock.Any()).
		Return(&pb.CreateCredentialResponse{}, nil)

	err := manager.CreateCredential(context.Background(), testCredential)
	if err != nil {
		t.Fatalf("CreateCredential failed: %v", err)
	}
}

func TestGetCredentialByID_Success(t *testing.T) {
	manager, ctrl, mockClient := setup(t)
	defer ctrl.Finish()

	resp := &pb.GetCredentialByIDResponse{}
	credProto := &pb.Credential{}
	credProto.SetId(testCredential.ID)
	credProto.SetUserId(testCredential.UserID)
	credProto.SetTitle(testCredential.Title)
	credProto.SetLogin(testCredential.Login)
	credProto.SetPassword(testCredential.Password)
	credProto.SetMetadata(testCredential.Metadata)
	credProto.SetCreatedAt(timestamppb.New(testCredential.CreatedAt))
	credProto.SetUpdatedAt(timestamppb.New(testCredential.UpdatedAt))
	resp.SetCredential(credProto)

	mockClient.EXPECT().
		GetCredentialByID(gomock.Any(), gomock.Any()).
		Return(resp, nil)

	cred, err := manager.GetCredentialByID(context.Background(), testCredential.ID)
	if err != nil {
		t.Fatalf("GetCredentialByID failed: %v", err)
	}
	if cred.ID != testCredential.ID {
		t.Errorf("Expected ID %s, got %s", testCredential.ID, cred.ID)
	}
}

func TestGetCredentials_Success(t *testing.T) {
	manager, ctrl, mockClient := setup(t)
	defer ctrl.Finish()

	credProto := &pb.Credential{}
	credProto.SetId(testCredential.ID)
	credProto.SetUserId(testCredential.UserID)
	credProto.SetTitle(testCredential.Title)
	credProto.SetLogin(testCredential.Login)
	credProto.SetPassword(testCredential.Password)
	credProto.SetMetadata(testCredential.Metadata)
	credProto.SetCreatedAt(timestamppb.New(testCredential.CreatedAt))
	credProto.SetUpdatedAt(timestamppb.New(testCredential.UpdatedAt))

	resp := &pb.GetCredentialsResponse{}
	resp.SetCredentials([]*pb.Credential{credProto})

	mockClient.EXPECT().
		GetCredentials(gomock.Any(), gomock.Any()).
		Return(resp, nil)

	creds, err := manager.GetCredentials(context.Background())
	if err != nil {
		t.Fatalf("GetCredentialsByUserID failed: %v", err)
	}
	if len(creds) != 1 {
		t.Errorf("Expected 1 credential, got %d", len(creds))
	}
}

func TestUpdateCredential_Success(t *testing.T) {
	manager, ctrl, mockClient := setup(t)
	defer ctrl.Finish()

	mockClient.EXPECT().
		UpdateCredential(gomock.Any(), gomock.Any()).
		Return(&pb.UpdateCredentialResponse{}, nil)

	err := manager.UpdateCredential(context.Background(), testCredential)
	if err != nil {
		t.Fatalf("UpdateCredential failed: %v", err)
	}
}

func TestDeleteCredential_Success(t *testing.T) {
	manager, ctrl, mockClient := setup(t)
	defer ctrl.Finish()

	mockClient.EXPECT().
		DeleteCredential(gomock.Any(), gomock.Any()).
		Return(&pb.DeleteCredentialResponse{}, nil)

	err := manager.DeleteCredential(context.Background(), testCredential.ID)
	if err != nil {
		t.Fatalf("DeleteCredential failed: %v", err)
	}
}
