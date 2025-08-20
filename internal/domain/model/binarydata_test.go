package model_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

func TestBinaryData_FormFields(t *testing.T) {
	data := &model.BinaryData{
		Title:      "File",
		Metadata:   "meta",
		ClientPath: "/path/to/file",
		UpdatedAt:  time.Now(),
	}

	fields := data.FormFields()
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(fields))
	}

	if fields[0].Label != "Title" || fields[0].Value != "File" {
		t.Errorf("Title field mismatch: %+v", fields[0])
	}
	if fields[1].Label != "Metadata" || fields[1].Value != "meta" {
		t.Errorf("Metadata field mismatch: %+v", fields[1])
	}
	if fields[1].InputType != "multiline" {
		t.Errorf("expected multiline type for Metadata, got %s", fields[1].InputType)
	}
	if fields[2].Label != "Client Path" || fields[2].Value != "/path/to/file" {
		t.Errorf("ClientPath field mismatch: %+v", fields[1])
	}

}

func TestBinaryData_UpdateFromFields(t *testing.T) {
	data := &model.BinaryData{}

	fields := []forms.FormField{
		{Value: "New Title"},
		{Value: "New meta"},
		{Value: "/tmp/file"},
		{Value: "UpdatedAt"},
	}

	err := data.UpdateFromFields(fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Title != "New Title" || data.ClientPath != "/tmp/file" || data.Metadata != "New meta" {
		t.Errorf("binary data not updated correctly: %+v", data)
	}

	if time.Since(data.UpdatedAt) > time.Second {
		t.Errorf("UpdatedAt not set correctly: %v", data.UpdatedAt)
	}
}

func TestBinaryData_UpdateFromFields_Errors(t *testing.T) {
	data := &model.BinaryData{}

	fields := []forms.FormField{{Value: "only one"}}
	err := data.UpdateFromFields(fields)
	if err == nil || !strings.Contains(err.Error(), "unexpected number") {
		t.Errorf("expected error on wrong field count, got %v", err)
	}

	fields = []forms.FormField{
		{Value: " "}, {Value: "/tmp/file"}, {Value: "meta"}, {Value: "UpdatedAt"},
	}
	err = data.UpdateFromFields(fields)
	if err == nil || !strings.Contains(err.Error(), "title cannot be empty") {
		t.Errorf("expected error on empty title, got %v", err)
	}
}

func TestBinaryData_GetSetID(t *testing.T) {
	data := &model.BinaryData{}
	data.SetID("123")
	if data.GetID() != "123" {
		t.Errorf("GetID returned %s, want 123", data.GetID())
	}
}
