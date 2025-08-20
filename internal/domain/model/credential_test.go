package model_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

func TestCredential_FormFields(t *testing.T) {
	cred := &model.Credential{
		Title:     "Gmail",
		Login:     "user@gmail.com",
		Password:  "secret",
		Metadata:  "notes",
		UpdatedAt: time.Now(),
	}

	fields := cred.FormFields()
	if len(fields) != 5 {
		t.Fatalf("expected 5 fields, got %d", len(fields))
	}

	if fields[0].Label != "Title" || fields[0].Value != "Gmail" {
		t.Errorf("Title field mismatch: %+v", fields[0])
	}
	if fields[2].InputType != "password" {
		t.Errorf("expected password type, got %s", fields[2].InputType)
	}
	if fields[3].InputType != "multiline" {
		t.Errorf("expected multiline type, got %s", fields[3].InputType)
	}
}

func TestCredential_UpdateFromFields(t *testing.T) {
	cred := &model.Credential{}

	fields := []forms.FormField{
		{Value: "NewTitle"},
		{Value: "newlogin"},
		{Value: "newpass"},
		{Value: "metadata text"},
		{Value: "UpdatedAt"},
	}

	err := cred.UpdateFromFields(fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cred.Title != "NewTitle" || cred.Login != "newlogin" || cred.Password != "newpass" || cred.Metadata != "metadata text" {
		t.Errorf("credential not updated correctly: %+v", cred)
	}

	if time.Since(cred.UpdatedAt) > time.Second {
		t.Errorf("UpdatedAt not set correctly: %v", cred.UpdatedAt)
	}
}

func TestCredential_UpdateFromFields_Errors(t *testing.T) {
	cred := &model.Credential{}

	// Ошибочное количество полей
	fields := []forms.FormField{{Value: "only one"}}
	err := cred.UpdateFromFields(fields)
	if err == nil || !strings.Contains(err.Error(), "unexpected number") {
		t.Errorf("expected error on wrong field count, got %v", err)
	}

	// Пустой Title
	fields = []forms.FormField{
		{Value: " "}, {Value: "login"}, {Value: "pass"}, {Value: "meta"}, {Value: "UpdatedAt"},
	}
	err = cred.UpdateFromFields(fields)
	if err == nil || !strings.Contains(err.Error(), "title cannot be empty") {
		t.Errorf("expected error on empty title, got %v", err)
	}
}

func TestCredential_GetSetID(t *testing.T) {
	cred := &model.Credential{}
	cred.SetID("123")
	if cred.GetID() != "123" {
		t.Errorf("GetID returned %s, want 123", cred.GetID())
	}
}
