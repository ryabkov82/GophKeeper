package model_test

import (
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

func TestTextData_GetSetID(t *testing.T) {
	tdata := &model.TextData{}
	tdata.SetID("abc123")
	if tdata.GetID() != "abc123" {
		t.Errorf("GetID returned %s, want abc123", tdata.GetID())
	}
}
