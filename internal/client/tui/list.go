package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// initDataView инициализирует просмотр данных
func initDataView(m Model) Model {
	m.currentState = "view_data"
	// Здесь будет инициализация списка данных
	return m
}

// updateViewData обрабатывает сообщения при просмотре данных
func updateViewData(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.listCursor > 0 {
				m.listCursor--
			}
		case "down", "j":
			if m.listCursor < len(m.listItems)-1 {
				m.listCursor++
			}
		case "enter":
			if len(m.listItems) > 0 {
				//selected := m.listItems[m.listCursor]
				// Загружаем полные данные по selected.ID и переходим в режим просмотра/редактирования
				//return loadAndShowItem(m, selected.ID)
			}
		case "tab":
			// Перейти в состояние добавления новой записи
			m.currentState = "edit_new"
			m.editEntity = newEmptyEntity(m.currentType) // функция создаёт пустую структуру Credential
			m.inputs = initFormInputs(m.editEntity.(forms.FormEntity))
			m.focusedInput = 0
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
	title := lipgloss.NewStyle().Bold(true).Render("Список данных:")

	b.WriteString(title + "\n\n")

	for i, item := range m.listItems {
		cursor := "  "
		if i == m.listCursor {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, item.Title))
	}

	b.WriteString("\n↑/↓: навигация • Enter: просмотр • Ins: добавить новую запись • Esc: назад")

	return b.String()
}

func newEmptyEntity(dataType contracts.DataType) interface{} {
	switch dataType {
	case contracts.TypeCredentials:
		return &model.Credential{}
	/*
		case contracts.TypeNotes:
			return &model.Note{} // если у тебя есть такой тип
		case contracts.TypeFiles:
			return &model.File{} // и т.д.
		case contracts.TypeCards:
			return &model.Card{}
	*/
	default:
		return nil
	}
}
