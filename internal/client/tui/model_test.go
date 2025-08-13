package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
)

// --- Фейковые сервисы ---
type fakeCredentialService struct{}

func (f *fakeCredentialService) List(ctx context.Context) ([]contracts.ListItem, error) {
	return []contracts.ListItem{
		{ID: "1", Title: "Cred1"},
		{ID: "2", Title: "Cred2"},
	}, nil
}
func (f *fakeCredentialService) Get(ctx context.Context, id string) (interface{}, error) {
	return &contracts.ListItem{ID: id, Title: "title"}, nil
}
func (f *fakeCredentialService) Create(ctx context.Context, v interface{}) error { return nil }
func (f *fakeCredentialService) Update(ctx context.Context, id string, v interface{}) error {
	return nil
}
func (f *fakeCredentialService) Delete(ctx context.Context, id string) error { return nil }

// --- Тесты ---

func TestUpdate_ListLoadedMsg(t *testing.T) {
	m := Model{
		currentState: "list",
		currentType:  contracts.TypeCredentials,
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeCredentials: &fakeCredentialService{}},
		ctx: context.Background(),
	}

	// Передаем listLoadedMsg
	msg := listLoadedMsg{
		items: []contracts.ListItem{{ID: "1", Title: "Cred1"}},
	}
	newModel, _ := m.Update(msg)

	m2 := newModel.(Model)
	if len(m2.listItems) != 1 || m2.listItems[0].Title != "Cred1" {
		t.Errorf("expected listItems to contain Cred1, got %+v", m2.listItems)
	}
	if m2.listCursor != 0 {
		t.Errorf("expected listCursor=0, got %d", m2.listCursor)
	}
}

func TestUpdate_ErrMsg(t *testing.T) {
	m := Model{
		currentState: "list",
		currentType:  contracts.TypeCredentials,
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeCredentials: &fakeCredentialService{}},
		ctx: context.Background(),
	}

	err := errors.New("some error")
	msg := errMsg{err: err}
	newModel, _ := m.Update(msg)
	m2 := newModel.(Model)
	if m2.listErr == nil || m2.listErr.Error() != "some error" {
		t.Errorf("expected listErr=some error, got %+v", m2.listErr)
	}
}

func TestView(t *testing.T) {
	m := Model{currentState: "menu"}
	view := m.View()
	if view == "" {
		t.Error("expected View() to return non-empty string for menu")
	}

	m.currentState = "unknown"
	view = m.View()
	if view != "" {
		t.Errorf("expected empty string for unknown state, got %q", view)
	}
}

func TestLoadList(t *testing.T) {
	m := Model{
		currentType: contracts.TypeCredentials,
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeCredentials: &fakeCredentialService{}},
		ctx: context.Background(),
	}
	cmd := m.loadList()
	msg := cmd()
	lm, ok := msg.(listLoadedMsg)
	if !ok {
		t.Fatalf("expected ListLoadedMsg, got %T", msg)
	}
	if len(lm.items) != 2 {
		t.Errorf("expected 2 items, got %d", len(lm.items))
	}
}

func TestExtractFields(t *testing.T) {
	m := Model{
		widgets: []formWidget{
			{
				isTextarea: false,
				input:      textinput.New(),
				field:      forms.FormField{Label: "Login"},
			},
			{
				isTextarea: true,
				textarea:   textarea.New(),
				field:      forms.FormField{Label: "Notes"},
			},
		},
	}
	m.widgets[0].input.SetValue("user1")
	m.widgets[1].textarea.SetValue("some notes")

	fields := m.ExtractFields()
	if fields[0].Value != "user1" {
		t.Errorf("expected Login='user1', got %s", fields[0].Value)
	}
	if fields[1].Value != "some notes" {
		t.Errorf("expected Notes='some notes', got %s", fields[1].Value)
	}
}
