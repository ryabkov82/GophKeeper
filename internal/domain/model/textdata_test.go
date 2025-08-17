package model_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

func TestTextData_FormFields(t *testing.T) {
	tdata := &model.TextData{
		Title:    "Notes",
		Content:  []byte("This is the note content."),
		Metadata: "Additional info",
	}

	fields := tdata.FormFields()
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}

	if fields[0].Label != "Title" || fields[0].Value != "Notes" {
		t.Errorf("Title field mismatch: %+v", fields[0])
	}
	if fields[1].Label != "Content" || fields[1].Value != "This is the note content." {
		t.Errorf("Content field mismatch: %+v", fields[1])
	}
	if fields[2].Label != "Metadata" || fields[2].Value != "Additional info" {
		t.Errorf("Metadata field mismatch: %+v", fields[2])
	}

	if fields[0].InputType != "text" {
		t.Errorf("expected text type for Title, got %s", fields[0].InputType)
	}
	if fields[1].InputType != "multiline" {
		t.Errorf("expected multiline type for Content, got %s", fields[1].InputType)
	}
	if fields[2].InputType != "multiline" {
		t.Errorf("expected multiline type for Metadata, got %s", fields[2].InputType)
	}
}

func TestTextData_UpdateFromFields(t *testing.T) {
	tdata := &model.TextData{}

	fields := []forms.FormField{
		{Value: "New Title"},
		{Value: "New content"},
		{Value: "Meta info"},
	}

	err := tdata.UpdateFromFields(fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tdata.Title != "New Title" {
		t.Errorf("Title not updated correctly, got %s", tdata.Title)
	}
	if string(tdata.Content) != "New content" {
		t.Errorf("Content not updated correctly, got %s", tdata.Content)
	}
	if tdata.Metadata != "Meta info" {
		t.Errorf("Metadata not updated correctly, got %s", tdata.Metadata)
	}

	if time.Since(tdata.UpdatedAt) > time.Second {
		t.Errorf("UpdatedAt not set correctly: %v", tdata.UpdatedAt)
	}
}

func TestTextData_UpdateFromFields_Errors(t *testing.T) {
	tdata := &model.TextData{}

	// Ошибочное количество полей
	fields := []forms.FormField{{Value: "only one"}}
	err := tdata.UpdateFromFields(fields)
	if err == nil || !strings.Contains(err.Error(), "unexpected number") {
		t.Errorf("expected error on wrong field count, got %v", err)
	}

	// Пустой Title
	fields = []forms.FormField{
		{Value: " "}, {Value: "content"}, {Value: "meta"},
	}
	err = tdata.UpdateFromFields(fields)
	if err == nil || !strings.Contains(err.Error(), "title cannot be empty") {
		t.Errorf("expected error on empty title, got %v", err)
	}
}

func TestTextData_GetSetID(t *testing.T) {
	tdata := &model.TextData{}
	tdata.SetID("abc123")
	if tdata.GetID() != "abc123" {
		t.Errorf("GetID returned %s, want abc123", tdata.GetID())
	}
}
