package model_test

import (
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

func TestCredential_GetSetID(t *testing.T) {
	cred := &model.Credential{}
	cred.SetID("123")
	if cred.GetID() != "123" {
		t.Errorf("GetID returned %s, want 123", cred.GetID())
	}
}
