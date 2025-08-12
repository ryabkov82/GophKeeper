package tui

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

func initEditForm(m Model) Model {
	//m.currentState = "edit"
	m.editErr = nil

	if entity, ok := m.editEntity.(forms.FormEntity); ok {
		m.editFields = entity.FormFields()
		m.inputs = initFormInputsFromFields(entity)
		m.focusedInput = 0
	} else {
		m.editErr = fmt.Errorf("entity does not implement FormEntity")
	}

	return m
}

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
			return updateInputFocus(m), nil

		case "esc":
			// Отмена редактирования/создания, возвращаемся в список
			m.currentState = "list"
			m.editEntity = nil
			m.inputs = nil
			return m, nil

		case "ctrl+s":
			// Сохраняем данные из формы
			saveEdit(m)
		case "enter":
			// Если последнее поле — сохраняем, иначе переходим дальше
			if m.focusedInput == len(m.inputs)-1 {
				return saveEdit(m)
			} else {
				m.focusedInput++
				return m, m.inputs[m.focusedInput].Focus()
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
	return m, cmd

}

func renderEditForm(m Model) string {
	var b strings.Builder

	title := "Редактирование записи"
	if m.currentState == "edit_new" {
		title = "Добавление новой записи"
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(title) + "\n\n")

	for i, input := range m.inputs {
		label := m.editFields[i].Label + ": "
		if i == m.focusedInput {
			label = activeFieldStyle.Render(label)
		} else {
			label = inactiveFieldStyle.Render(label)
		}

		b.WriteString(label + input.View() + "\n")
	}

	if m.editErr != nil {
		b.WriteString("\n" + errorStyle.Render("Ошибка: "+m.editErr.Error()))
	}

	b.WriteString("\n" + hintStyle.Render(
		"Esc: Отмена • Enter/Ctrl+S: Сохранить • Tab: Следующее поле\n",
	))

	return b.String()
}

// saveEdit — вынесена логика сохранения в отдельную функцию
func saveEdit(m Model) (Model, tea.Cmd) {
	m = updateEditEntityFromInputs(m)
	if m.editErr != nil {
		return m, nil
	}

	var err error
	if m.currentState == "edit_new" {
		err = m.services[m.currentType].Create(m.ctx, m.editEntity)
	} else {
		if idGetter, ok := m.editEntity.(forms.Identifiable); ok {
			id := idGetter.GetID()
			err = m.services[m.currentType].Update(m.ctx, id, m.editEntity)
		} else {
			m.editErr = errors.New("missing entity ID")
			return m, nil
		}
	}

	if err != nil {
		m.editErr = err
		return m, nil
	}

	// Успешно
	m.listItems, _ = m.services[m.currentType].List(m.ctx)
	m.currentState = "list"
	m.editEntity = nil
	m.inputs = nil
	m.focusedInput = 0
	return m, nil
}

// updateEditEntityFromInputs обновляет m.editEntity значениями из m.inputs.
// Возвращает обновлённую модель (m.editErr заполняется при ошибке).
func updateEditEntityFromInputs(m Model) Model {
	// 1) Приведение к forms.FormEntity
	fe, ok := m.editEntity.(forms.FormEntity)
	if !ok {
		m.editErr = fmt.Errorf("entity does not implement forms.FormEntity")
		return m
	}

	// Используем поля из модели (инициализированные один раз при входе в форму)
	fields := extractFieldsFromInputs(m.editFields, m.inputs)

	// 4) Валидация и обновление самой сущности (реализовано в доменной модели)
	if err := fe.UpdateFromFields(fields); err != nil {
		m.editErr = err
		return m
	}

	// 5) Опционально — если нужно, можем вернуть обновлённую сущность в m.editEntity.
	// Обычно fe ссылается на тот же указатель, что и m.editEntity, поэтому этого не требуется,
	// но для чёткости присвоим:
	m.editEntity = fe
	m.editErr = nil
	return m
}
