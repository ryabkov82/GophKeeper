package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// initListForm инициализирует просмотр данных
func initListForm(m Model, dataType contracts.DataType) Model {
	m.currentState = "list"
	m.currentType = dataType
	m.listItems = nil // очищаем старый список
	m.listCursor = 0
	m.editEntity = nil // сбрасываем редактируемую сущность
	m.inputs = nil     // очищаем поля ввода формы
	m.focusedInput = 0
	m.widgets = nil
	m.listErr = nil

	return m
}

// updateViewData обрабатывает сообщения при просмотре данных
func updateViewData(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "shift+tab":
			if m.listCursor > 0 {
				m.listCursor--
			}
		case "down", "tab":
			if m.listCursor < len(m.listItems)-1 {
				m.listCursor++
			}
		case "enter":
			if len(m.listItems) > 0 {
				selected := m.listItems[m.listCursor]
				// Загружаем полные данные по selected.ID и переходим в режим просмотра/редактирования
				return loadAndShowItem(m, selected.ID)
			}
		case "ctrl+n":
			// Перейти в состояние добавления новой записи
			m.currentState = "edit_new"
			m.editEntity = newEmptyEntity(m.currentType) // функция создаёт пустую структуру соответствующего типа
			m = initEditForm(m)
		case "ctrl+d":
			// Удаляем выбранную сущность
			if len(m.listItems) > 0 {
				selected := m.listItems[m.listCursor]
				err := m.services[m.currentType].Delete(m.ctx, selected.ID)
				if err != nil {
					m.listErr = fmt.Errorf("failed to delete item: %w", err)
				} else {
					// Обновляем список после удаления
					m = initListForm(m, m.currentType)
					return m, m.loadList()
				}
			}
		case "esc":
			// Возврат в главное меню
			m.currentState = "menu"
		}
	}
	return m, nil
}

// renderViewData отображает список данных
func renderList(m Model) string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Render("Список данных " + m.currentType.String() + ":")

	b.WriteString(title + "\n\n")

	for i, item := range m.listItems {
		cursor := "  "
		if i == m.listCursor {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, item.Title))
	}

	if m.listErr != nil {
		b.WriteString("\n" + errorStyle.Render("Ошибка: "+m.listErr.Error()))
	}

	b.WriteString("\n" + hintStyle.Render(
		"↑/↓: навигация • Enter: просмотр • Ctrl+N: добавить новую запись • Esc: назад",
	))

	return b.String()
}

func newEmptyEntity(dataType contracts.DataType) interface{} {
	switch dataType {
	case contracts.TypeCredentials:
		return &model.Credential{}
	case contracts.TypeCards:
		return &model.BankCard{}
		/*
			case contracts.TypeNotes:
				return &model.Note{} // если у тебя есть такой тип
			case contracts.TypeFiles:
				return &model.File{} // и т.д.
		*/
	default:
		return nil
	}
}

func loadAndShowItem(m Model, id string) (Model, tea.Cmd) {
	entity, err := m.services[m.currentType].Get(m.ctx, id)
	if err != nil {
		m.listErr = fmt.Errorf("failed to load item: %w", err)
		return m, nil
	}

	m.editEntity = entity
	m.currentState = "edit"
	m = initEditForm(m)

	return m, nil
}
