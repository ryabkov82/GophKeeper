package model_test

import (
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

func TestBinaryData_GetSetID(t *testing.T) {
	data := &model.BinaryData{}
	data.SetID("123")
	if data.GetID() != "123" {
		t.Errorf("GetID returned %s, want 123", data.GetID())
	}
}
