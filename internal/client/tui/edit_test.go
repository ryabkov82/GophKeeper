package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
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

func TestUpdateEdit_EnterMovesNextAndEditsTextarea(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "A", Value: "a"},                 // обычное поле
		{Label: "Notes", InputType: "multiline"}, // textarea
	}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 20)
	m := Model{
		editEntity:   entity,
		widgets:      widgets,
		focusedInput: focused,
	}

	// Enter на обычном поле -> фокус смещается на следующее
	m1, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m1.focusedInput != 1 {
		t.Fatalf("expected focus=1 after Enter on input, got %d", m1.focusedInput)
	}

	// Enter на textarea -> остаёмся на том же поле, внутри textarea добавляется перенос
	before := m1.widgets[1].textarea.Value()
	m2, _ := updateEdit(m1, tea.KeyMsg{Type: tea.KeyEnter})
	after := m2.widgets[1].textarea.Value()
	if m2.focusedInput != 1 {
		t.Fatalf("expected focus stay on textarea, got %d", m2.focusedInput)
	}
	if len(after) != len(before)+1 {
		t.Fatalf("expected textarea to receive newline on Enter (len+1), got before=%d after=%d", len(before), len(after))
	}
}

func TestUpdateEdit_UpDownNavigation_WrapAndTextareaStay(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "A"}, {Label: "B"}, {Label: "C"},
	}}
	widgets, _ := initFormInputsFromFields(entity.FormFields(), 20)
	m := Model{editEntity: entity, widgets: widgets, focusedInput: 1}

	// Down -> 2
	m1, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyDown})
	if m1.focusedInput != 2 {
		t.Fatalf("down expected 2, got %d", m1.focusedInput)
	}
	// Down (wrap) -> 0
	m2, _ := updateEdit(m1, tea.KeyMsg{Type: tea.KeyDown})
	if m2.focusedInput != 0 {
		t.Fatalf("down wrap expected 0, got %d", m2.focusedInput)
	}
	// Up (wrap) -> 2
	m3, _ := updateEdit(m2, tea.KeyMsg{Type: tea.KeyUp})
	if m3.focusedInput != 2 {
		t.Fatalf("up wrap expected 2, got %d", m3.focusedInput)
	}

	// Теперь проверим textarea: up/down НЕ двигают фокус, а отдаются самой textarea
	entityTA := &fakeForm{fields: []forms.FormField{
		{Label: "Notes", InputType: "multiline", Value: "l1\nl2"},
	}}
	wTA, focused := initFormInputsFromFields(entityTA.FormFields(), 20)
	mTA := Model{editEntity: entityTA, widgets: wTA, focusedInput: focused}

	mTA1, _ := updateEdit(mTA, tea.KeyMsg{Type: tea.KeyDown})
	if mTA1.focusedInput != 0 {
		t.Fatalf("textarea: expected focus stay=0 on down, got %d", mTA1.focusedInput)
	}
	mTA2, _ := updateEdit(mTA1, tea.KeyMsg{Type: tea.KeyUp})
	if mTA2.focusedInput != 0 {
		t.Fatalf("textarea: expected focus stay=0 on up, got %d", mTA2.focusedInput)
	}
}

func TestUpdateEdit_CtrlBPasswordToggle_FirstOnly(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "Pwd1", Value: "s1", InputType: "password"},
		{Label: "Pwd2", Value: "s2", InputType: "password"},
		{Label: "User", Value: "u"},
	}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 20)
	m := Model{editEntity: entity, widgets: widgets, focusedInput: focused}

	// До переключения: оба пароля скрыты
	if widgets[0].input.EchoMode != textinput.EchoPassword || widgets[1].input.EchoMode != textinput.EchoPassword {
		t.Fatalf("expected both passwords to start hidden")
	}

	// Ctrl+B -> переключается ТОЛЬКО первый встретившийся пароль
	m1, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyCtrlB})
	if m1.widgets[0].input.EchoMode != textinput.EchoNormal {
		t.Fatalf("first password should be visible after Ctrl+B")
	}
	if m1.widgets[1].input.EchoMode != textinput.EchoPassword {
		t.Fatalf("second password should remain hidden")
	}

	// Повторный Ctrl+B -> возвращаем первый к EchoPassword
	m2, _ := updateEdit(m1, tea.KeyMsg{Type: tea.KeyCtrlB})
	if m2.widgets[0].input.EchoMode != textinput.EchoPassword {
		t.Fatalf("first password should toggle back to hidden")
	}
}

func TestUpdateEdit_F2OpensFullscreen_WhenAllowed(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{
		{Label: "Notes", InputType: "multiline", Fullscreen: true, Value: "abc"},
	}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 60)
	m := Model{
		editEntity:   entity,
		widgets:      widgets,
		focusedInput: focused,
		currentState: "edit",
	}

	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyF2})
	if m2.currentState != "fullscreen_edit" {
		t.Fatalf("expected fullscreen_edit state after F2, got %q", m2.currentState)
	}
	if m2.prevState != "edit" {
		t.Fatalf("expected prevState=edit, got %q", m2.prevState)
	}
	if m2.fullscreenWidget == nil {
		t.Fatalf("expected fullscreenWidget to be set")
	}
}

func TestUpdateEdit_F2Ignored_OnSimpleInput(t *testing.T) {
	entity := &fakeForm{fields: []forms.FormField{{Label: "A", Value: "a"}}}
	widgets, focused := initFormInputsFromFields(entity.FormFields(), 20)
	m := Model{editEntity: entity, widgets: widgets, focusedInput: focused, currentState: "edit"}

	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyF2})
	if m2.currentState != "edit" {
		t.Fatalf("expected state to remain 'edit' for non-textarea on F2, got %q", m2.currentState)
	}
}

func TestUpdateEdit_NoWidgets_NoCrash(t *testing.T) {
	m := Model{widgets: nil, focusedInput: 0, currentState: "edit"}
	if _, cmd := updateEdit(m, tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatalf("expected no cmd when no widgets on Enter")
	}
	if m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyDown}); m2.focusedInput != 0 {
		t.Fatalf("expected focus unchanged when no widgets on Down")
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

func TestUpdateEdit_CtrlU_OpensUpload_ForFiles(t *testing.T) {
	bd := &model.BinaryData{Title: "old", ClientPath: "/tmp/file.bin"} // entity
	m := Model{
		currentState: "edit",
		currentType:  contracts.TypeFiles,
		editEntity:   bd,
		widgets:      nil,
		focusedInput: 0,
	}

	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	if m2.currentState != "file_transfer" {
		t.Fatalf("expected to enter file_transfer on Ctrl+U, got %q", m2.currentState)
	}
	if m2.transfer.mode != modeUpload {
		t.Fatalf("expected modeUpload, got %v", m2.transfer.mode)
	}
	if m2.transfer.data == nil {
		t.Fatalf("expected transfer.data to be set")
	}
}

func TestUpdateEdit_CtrlU_Ignored_ForNonFiles(t *testing.T) {
	m := Model{
		currentState: "edit",
		currentType:  contracts.TypeCards, // не файлы
		editEntity:   &fakeForm{},         // любая сущность
		widgets:      nil,
		focusedInput: 0,
	}
	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	if m2.currentState != "edit" {
		t.Fatalf("expected state stay 'edit' for non-files Ctrl+U, got %q", m2.currentState)
	}
}

func TestUpdateEdit_CtrlD_Download_ErrWhenClientPathEmpty(t *testing.T) {
	// ClientPath пуст → нельзя в download
	bd := &model.BinaryData{ID: "id-1", ClientPath: ""}
	m := Model{
		currentState: "edit",
		currentType:  contracts.TypeFiles,
		editEntity:   bd,
		widgets:      nil,
		focusedInput: 0,
	}

	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	if m2.currentState != "edit" {
		t.Fatalf("expected stay in 'edit' when ClientPath is empty, got %q", m2.currentState)
	}
	if m2.editErr == nil {
		t.Fatalf("expected editErr when ClientPath is empty")
	}
	if got := m2.editErr.Error(); !strings.Contains(strings.ToLower(got), "скачив") {
		t.Fatalf("unexpected error message: %v", got)
	}
}

func TestUpdateEdit_CtrlD_Download_ErrWhenIDEmpty(t *testing.T) {
	// ClientPath задан, но ID пуст → нельзя в download
	bd := &model.BinaryData{ID: "", ClientPath: "/tmp/file.bin"}
	m := Model{
		currentState: "edit",
		currentType:  contracts.TypeFiles,
		editEntity:   bd,
		widgets:      nil,
		focusedInput: 0,
	}

	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	if m2.currentState != "edit" {
		t.Fatalf("expected stay in 'edit' when ID is empty, got %q", m2.currentState)
	}
	if m2.editErr == nil {
		t.Fatalf("expected editErr when ID is empty")
	}
	if got := m2.editErr.Error(); !strings.Contains(strings.ToLower(got), "сначала сохраните") {
		t.Fatalf("unexpected error message: %v", got)
	}
}

func TestUpdateEdit_CtrlD_OpensDownload_ForFiles(t *testing.T) {
	// Валидный кейс: и ClientPath, и ID заданы → входим в download
	bd := &model.BinaryData{ID: "id-123", ClientPath: "/tmp/file.bin"}
	m := Model{
		currentState: "edit",
		currentType:  contracts.TypeFiles,
		editEntity:   bd,
		widgets:      nil,
		focusedInput: 0,
	}

	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	if m2.currentState != "file_transfer" {
		t.Fatalf("expected to enter file_transfer on Ctrl+D, got %q", m2.currentState)
	}
	if m2.transfer.mode != modeDownload {
		t.Fatalf("expected modeDownload, got %v", m2.transfer.mode)
	}
	if m2.transfer.data == nil || m2.transfer.data.ID != "id-123" {
		t.Fatalf("expected transfer.data.ID to be 'id-123'")
	}
}

func TestUpdateEdit_CtrlD_Ignored_ForNonFiles(t *testing.T) {
	m := Model{
		currentState: "edit",
		currentType:  contracts.TypeNotes, // любой тип кроме файлов
		editEntity:   &fakeForm{},
		widgets:      nil,
		focusedInput: 0,
	}
	m2, _ := updateEdit(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	if m2.currentState != "edit" {
		t.Fatalf("expected state stay 'edit' for non-files Ctrl+D, got %q", m2.currentState)
	}
}
