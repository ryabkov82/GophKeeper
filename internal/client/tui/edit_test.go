package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/stretchr/testify/assert"
)

// --- Фейковая форма ---
type fakeForm struct {
	fields []forms.FormField
	ID     string
}

func (f *fakeForm) FormFields() []forms.FormField {
	return f.fields
}

func (f *fakeForm) UpdateFromFields(fields []forms.FormField) error {
	if len(fields) != len(f.fields) {
		return errors.New("wrong number of fields")
	}
	f.fields = fields
	return nil
}

// Реализация forms.Identifiable
func (f *fakeForm) GetID() string {
	return f.ID
}

func (f *fakeForm) SetID(id string) {
	f.ID = id
}

// --- Фейковый DataService ---
type fakeEditDataService struct {
	created bool
	updated bool
}

func (f *fakeEditDataService) List(ctx context.Context) ([]contracts.ListItem, error) {
	return nil, nil
}
func (f *fakeEditDataService) Get(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}
func (f *fakeEditDataService) Create(ctx context.Context, v interface{}) error {
	f.created = true
	return nil
}
func (f *fakeEditDataService) Update(ctx context.Context, id string, v interface{}) error {
	f.updated = true
	return nil
}
func (f *fakeEditDataService) Delete(ctx context.Context, id string) error { return nil }

// --- Тесты ---

func TestInitEditForm(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "A", Value: "a"},
	}}
	m := Model{editEntity: entity}
	m2 := initEditForm(m)

	if m2.editErr != nil {
		t.Errorf("expected no editErr, got %v", m2.editErr)
	}
	if len(m2.widgets) != 1 {
		t.Errorf("expected 1 widget, got %d", len(m2.widgets))
	}

	// Тест с не-FormEntity
	m3 := Model{editEntity: struct{}{}}
	m3 = initEditForm(m3)
	if m3.editErr == nil {
		t.Errorf("expected editErr for non-FormEntity")
	}
}

func TestRenderEditForm(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "A", Value: "a"},
	}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 10)

	m := Model{
		editEntity:   entity,
		widgets:      widgets,
		focusedInput: focused,
		currentState: "edit_new",
	}

	output := renderEditForm(m)
	if output == "" {
		t.Errorf("expected non-empty render output")
	}
	assert.Contains(t, output, "Добавление новой записи")
	assert.Contains(t, output, "A:")
}

func TestRenderEditForm_ReadOnlyStyle(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{{Label: "A", Value: "a", ReadOnly: true}}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 10)
	m := Model{editEntity: entity, widgets: widgets, focusedInput: focused}
	output := renderEditForm(m)
	styled := readonlyStyle.Render("A: ")
	assert.Contains(t, output, styled)
}

func TestUpdateEdit_TabNavigation(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 10)
	m := Model{
		editEntity:   entity,
		widgets:      widgets,
		focusedInput: focused,
	}

	// Tab → должно переключить фокус
	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyTab})
	if m2.focusedInput != 1 {
		t.Errorf("expected focusedInput=1 after tab, got %d", m2.focusedInput)
	}

	// Shift+Tab → обратно
	m3, _ := updateEdit(m2, tea.KeyMsg{Type: tea.KeyShiftTab})
	if m3.focusedInput != 0 {
		t.Errorf("expected focusedInput=0 after shift+tab, got %d", m3.focusedInput)
	}
}

func TestUpdateEdit_SkipReadOnly(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "A", Value: "a", ReadOnly: true},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c", ReadOnly: true},
		{Label: "D", Value: "d"},
	}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 10)
	m := Model{editEntity: entity, widgets: widgets, focusedInput: focused}

	assert.Equal(t, 1, m.focusedInput)

	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 3, m2.focusedInput)

	m3, _ := updateEdit(m2, tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 1, m3.focusedInput)

	m4, _ := updateEdit(m3, tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, 3, m4.focusedInput)
}

func TestUpdateEdit_CtrlVPasswordToggle(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "Pwd", Value: "secret", InputType: "password"},
	}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 10)
	m := Model{
		editEntity:   entity,
		widgets:      widgets,
		focusedInput: focused,
	}

	// Ctrl+V → переключение видимости пароля
	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyCtrlV})
	w := m2.widgets[0]
	if w.input.EchoMode != textinput.EchoNormal && w.input.EchoMode != textinput.EchoPassword {
		t.Errorf("unexpected EchoMode after Ctrl+V")
	}
}

func TestUpdateEdit_EscCancels(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{{Label: "A", Value: "a"}}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 10)
	m := Model{
		editEntity:   entity,
		widgets:      widgets,
		focusedInput: focused,
		currentState: "edit",
	}

	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m2.currentState != "list" || m2.editEntity != nil || m2.inputs != nil {
		t.Errorf("expected cancel to reset editEntity and inputs")
	}
}

func TestSaveEdit_CreateAndUpdate(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{{Label: "A", Value: "a"}}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 10)
	fakeSvc := &fakeEditDataService{}

	// Новая сущность
	m := Model{
		editEntity:   entity,
		widgets:      widgets,
		focusedInput: focused,
		currentState: "edit_new",
		currentType:  contracts.TypeCredentials,
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeCredentials: fakeSvc,
		},
		ctx: context.Background(),
	}
	saveEdit(m)
	if !fakeSvc.created {
		t.Errorf("expected entity to be created")
	}

	// Существующая сущность
	fakeSvc2 := &fakeEditDataService{}
	m3 := Model{
		editEntity:   entity,
		widgets:      widgets,
		focusedInput: focused,
		currentState: "edit",
		currentType:  contracts.TypeCredentials,
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeCredentials: fakeSvc2,
		},
		ctx: context.Background(),
	}
	saveEdit(m3)
	if !fakeSvc2.updated {
		t.Errorf("expected entity to be updated")
	}
}
