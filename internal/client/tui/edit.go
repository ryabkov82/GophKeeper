package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func updateEdit(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := len(m.inputs)
			if s == 0 {
				return m, nil
			}

			// Навигация по полям формы
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusedInput--
				if m.focusedInput < 0 {
					m.focusedInput = s - 1
				}
			} else {
				m.focusedInput++
				if m.focusedInput >= s {
					m.focusedInput = 0
				}
			}
			return m, m.inputs[m.focusedInput].Focus()

		case "esc":
			// Отмена редактирования/создания, возвращаемся в список
			m.currentState = "view_data"
			m.editEntity = nil
			m.inputs = nil
			return m, nil

		case "ctrl+s", "enter":
			// Сохраняем данные из формы
			/*

				m = updateEditEntityFromInputs(m)

				var err error
				ctx := m.ctx

				if m.currentState == "edit_new" {
					err = m.services[m.currentType].Create(ctx, m.editEntity)
				} else if m.currentState == "edit" {
					// Предполагается, что id уже есть в m.editEntity
					id := getIDFromEntity(m.editEntity)
					err = m.services[m.currentType].Update(ctx, id, m.editEntity)
				}

				if err != nil {
					// Можно сохранить ошибку в модель и показать пользователю
					m.lastError = err
					return m, nil
				}

				// Успешно сохранили, обновляем список и возвращаемся к просмотру
				m.currentState = "list"
				m.editEntity = nil
				m.inputs = nil
				m.listItems, _ = m.services[m.currentType].List(ctx) // Обновить список после изменений
				m.listCursor = 0
				return m, nil
			*/
		}
	}

	// Обновляем активный input
	if len(m.inputs) > 0 && m.focusedInput >= 0 && m.focusedInput < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
		return m, cmd
	}

	return m, nil
}

func renderEditForm(m Model) string {
	var b strings.Builder

	title := "Редактирование записи"
	if m.currentState == "edit_new" {
		title = "Добавление новой записи"
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(title) + "\n\n")

	for i, input := range m.inputs {
		prefix := "  "
		if i == m.focusedInput {
			prefix = "> "
		}
		b.WriteString(prefix + input.View() + "\n")
	}

	b.WriteString("\n")
	b.WriteString("Esc: Отмена • Enter/Ctrl+S: Сохранить • Tab: Следующее поле\n")

	return b.String()
}

/*
func updateEditEntityFromInputs(m Model) Model {
	// Получить шаблон полей формы
	fields := forms.BuildFormFieldsForEntity(m.editEntity)

	// Обновить поля значениями из UI
	fields = extractFieldsFromInputs(fields, m.inputs)

	// Обновить сущность из заполненных полей
	err := m.editEntity.UpdateFromFields(fields)
	if err != nil {
		m.editErr = err
	} else {
		m.editErr = nil
	}

	return m
}
*/
