package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// --- Фейковый DataService ---
type fakeDataService struct {
	data map[string]interface{}
}

func (f *fakeDataService) List(ctx context.Context) ([]contracts.ListItem, error) {
	items := []contracts.ListItem{}
	for id, _ := range f.data {
		items = append(items, contracts.ListItem{ID: id, Title: id})
	}
	return items, nil
}

func (f *fakeDataService) Get(ctx context.Context, id string) (interface{}, error) {
	v, ok := f.data[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}
func (f *fakeDataService) Create(ctx context.Context, v interface{}) error            { return nil }
func (f *fakeDataService) Update(ctx context.Context, id string, v interface{}) error { return nil }
func (f *fakeDataService) Delete(ctx context.Context, id string) error {
	if _, ok := f.data[id]; ok {
		delete(f.data, id)
		return nil
	}
	return fmt.Errorf("item with ID %s not found", id)
}

// --- Тесты ---

func TestInitListForm(t *testing.T) {
	m := Model{
		currentState: "edit",
		listItems:    []contracts.ListItem{{ID: "1"}},
		listCursor:   2,
		editEntity:   struct{}{},
		inputs:       []textinput.Model{},
		widgets:      []formWidget{},
		listErr:      errors.New("err"),
	}
	m2 := initListForm(m, contracts.TypeCredentials)

	if m2.currentState != "list" {
		t.Errorf("expected currentState=list, got %s", m2.currentState)
	}
	if m2.currentType != contracts.TypeCredentials {
		t.Errorf("expected currentType=TypeCredentials, got %v", m2.currentType)
	}
	if len(m2.listItems) != 0 {
		t.Errorf("expected listItems to be empty")
	}
	if m2.listCursor != 0 {
		t.Errorf("expected listCursor=0, got %d", m2.listCursor)
	}
	if m2.editEntity != nil || m2.inputs != nil || m2.widgets != nil {
		t.Errorf("expected editEntity, inputs, widgets to be nil")
	}
	if m2.listErr != nil {
		t.Errorf("expected listErr to be nil")
	}
}

func TestUpdateViewDataNavigation(t *testing.T) {
	m := Model{
		listItems:   []contracts.ListItem{{ID: "1"}, {ID: "2"}, {ID: "3"}},
		listCursor:  1,
		currentType: contracts.TypeCredentials,
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeCredentials: &fakeDataService{data: map[string]interface{}{"1": &model.Credential{}, "2": &model.Credential{}}},
		},
		ctx: context.Background(),
	}

	// "up"
	m, _ = updateViewData(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.listCursor != 0 {
		t.Errorf("expected cursor=0, got %d", m.listCursor)
	}

	// "down"
	m, _ = updateViewData(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.listCursor != 1 {
		t.Errorf("expected cursor=1, got %d", m.listCursor)
	}

	// "esc"
	m, _ = updateViewData(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.currentState != "menu" {
		t.Errorf("expected currentState=menu, got %s", m.currentState)
	}
}

func TestRenderList(t *testing.T) {
	m := Model{
		currentType: contracts.TypeCredentials,
		listItems:   []contracts.ListItem{{ID: "1", Title: "A"}, {ID: "2", Title: "B"}},
		listCursor:  1,
		listErr:     errors.New("oops"),
	}

	output := renderList(m)
	if !strings.Contains(output, "A") || !strings.Contains(output, "B") {
		t.Errorf("expected output to contain item titles")
	}
	if !strings.Contains(output, "oops") {
		t.Errorf("expected output to contain error")
	}
}

func TestLoadAndShowItem(t *testing.T) {
	svc := &fakeDataService{
		data: map[string]interface{}{"1": &model.Credential{ID: "1"}},
	}
	m := Model{
		currentType: contracts.TypeCredentials,
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeCredentials: svc,
		},
		ctx: context.Background(),
	}

	m2, _ := loadAndShowItem(m, "1")
	if m2.currentState != "edit" {
		t.Errorf("expected currentState=edit, got %s", m2.currentState)
	}
	if m2.editEntity == nil {
		t.Errorf("expected editEntity to be set")
	}

	// Ошибка при несуществующем ID
	m3, _ := loadAndShowItem(m, "notfound")
	if m3.listErr == nil {
		t.Errorf("expected listErr to be set on failure")
	}
}

func TestUpdateViewData_DeleteItem(t *testing.T) {
	// Подготовка фейкового сервиса с данными
	fakeSvc := &fakeDataService{
		data: map[string]interface{}{
			"1": &model.Credential{ID: "1", Title: "Item 1"},
			"2": &model.Credential{ID: "2", Title: "Item 2"},
		},
	}

	// Инициализация модели
	m := initListForm(Model{
		currentState: "list",
		ctx:          context.Background(),
		services:     map[contracts.DataType]contracts.DataService{contracts.TypeCredentials: fakeSvc},
	}, contracts.TypeCredentials)

	// Обновляем список через loadList
	cmd := m.loadList()
	msg := cmd()                     // msg имеет тип tea.Msg
	updatedModel, _ := m.Update(msg) // updatedModel имеет тип tea.Model
	m = updatedModel.(Model)         // приведение типа к Model

	selectedID := m.listItems[m.listCursor].ID

	// Отправляем "ctrl+d" для удаления
	m, _ = updateViewData(m, tea.KeyMsg{Type: tea.KeyCtrlD})

	// Проверяем, что элемент удалён
	if _, exists := fakeSvc.data[selectedID]; exists {
		t.Errorf("expected item %s to be deleted", selectedID)
	}

	// Проверяем, что курсор корректно обновился
	if m.listCursor >= len(fakeSvc.data) {
		t.Errorf("listCursor out of bounds: %d", m.listCursor)
	}
}
