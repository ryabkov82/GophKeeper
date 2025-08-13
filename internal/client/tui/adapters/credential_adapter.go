package adapters

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// CredentialAdapter — адаптер, превращающий CredentialService в DataService
type CredentialAdapter struct {
	svc contracts.CredentialService
}

// NewCredentialAdapter создаёт адаптер
func NewCredentialAdapter(svc contracts.CredentialService) *CredentialAdapter {
	return &CredentialAdapter{svc: svc}
}

// List — возвращает список учётных данных в виде []ListItem
func (a *CredentialAdapter) List(ctx context.Context) ([]contracts.ListItem, error) {
	creds, err := a.svc.GetCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	items := make([]contracts.ListItem, 0, len(creds))
	for _, c := range creds {
		items = append(items, contracts.ListItem{
			ID:    c.ID,
			Title: c.Title,
		})
	}
	return items, nil
}

// Get — возвращает Credential по ID
func (a *CredentialAdapter) Get(ctx context.Context, id string) (interface{}, error) {
	cred, err := a.svc.GetCredentialByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	return cred, nil
}

// Create — создаёт новый Credential
func (a *CredentialAdapter) Create(ctx context.Context, v interface{}) error {
	cred, ok := v.(*model.Credential)
	if !ok {
		return fmt.Errorf("invalid type for Create: %T", v)
	}
	return a.svc.CreateCredential(ctx, cred)
}

// Update — обновляет Credential
func (a *CredentialAdapter) Update(ctx context.Context, id string, v interface{}) error {
	cred, ok := v.(*model.Credential)
	if !ok {
		return fmt.Errorf("invalid type for Update: %T", v)
	}
	if cred.ID != id {
		cred.ID = id
	}
	return a.svc.UpdateCredential(ctx, cred)
}

// Delete — удаляет Credential
func (a *CredentialAdapter) Delete(ctx context.Context, id string) error {
	return a.svc.DeleteCredential(ctx, id)
}
