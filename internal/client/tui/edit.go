package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

func initEditForm(m Model) Model {

	m.editErr = nil

	if entity, ok := m.editEntity.(forms.FormEntity); ok {
		formFields := entity.FormFields()
		m.widgets = initFormInputsFromFields(formFields)
		m.focusedInput = 0
	} else {
		m.editErr = fmt.Errorf("entity does not implement FormEntity")
	}

	return m
}

func updateEdit(m Model, msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Обработка глобальных горячих клавиш
		switch key {
		case "tab", "shift+tab":
			m = moveFocus(m, key == "tab")
			return focusField(m), nil

		case "up", "down":
			if len(m.widgets) == 0 {
				return m, nil
			}
			w := m.widgets[m.focusedInput]
			if w.isTextarea && strings.Contains(w.textarea.Value(), "\n") {
				var cmd tea.Cmd
				w.textarea, cmd = w.textarea.Update(msg)
				m.widgets[m.focusedInput] = w
				return m, cmd
			}
			m = moveFocus(m, key == "down")
			return focusField(m), nil

		case "esc":
			m.currentState = "list"
			m.editEntity = nil
			m.inputs = nil
			return m, nil

		case "ctrl+s":
			return saveEdit(m)

		case "enter":
			if len(m.widgets) == 0 {
				return m, nil
			}
			w := m.widgets[m.focusedInput]
			if w.isTextarea {
				var cmd tea.Cmd
				w.textarea, cmd = w.textarea.Update(msg)
				m.widgets[m.focusedInput] = w
				return m, cmd
			}
			m.focusedInput = (m.focusedInput + 1) % len(m.widgets)
			return focusField(m), nil

		case "ctrl+b":
			for i, w := range m.widgets {
				if !w.isTextarea && strings.ToLower(w.field.InputType) == "password" {
					if w.input.EchoMode == textinput.EchoPassword {
						w.input.EchoMode = textinput.EchoNormal
					} else {
						w.input.EchoMode = textinput.EchoPassword
					}
					m.widgets[i] = w
					break
				}
			}
			return m, nil
		}

		if len(m.widgets) == 0 {
			return m, nil
		}

		w := &m.widgets[m.focusedInput]

		// --- Обработка maskedInput ---
		if w.maskedInput != nil && w.maskedInput.Mask != "" {
			HandleMaskedInput(w, key, msg)
			return m, nil
		}

		if w.isTextarea {
			var cmd tea.Cmd
			w.textarea, cmd = w.textarea.Update(msg)
			return m, cmd
		}

		var cmd tea.Cmd
		w.input, cmd = w.input.Update(msg)
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

	// Обертка при рендере
	blockStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("15")).
		Padding(1, 1)

	for i, widget := range m.widgets {
		label := widget.field.Label + ": "
		if i == m.focusedInput {
			label = activeFieldStyle.Render(label)
		} else {
			label = inactiveFieldStyle.Render(label)
		}

		if widget.isTextarea {
			b.WriteString(label + "\n" + blockStyle.Render(widget.textarea.View()) + "\n")
		} else {
			b.WriteString(label + widget.input.View() + "\n")
		}
	}

	if m.editErr != nil {
		b.WriteString("\n" + errorStyle.Render("Ошибка: "+m.editErr.Error()))
	}

	b.WriteString("\n" + hintStyle.Render(
		"Esc: Отмена • Ctrl+S: Сохранить • Tab: Следующее поле • Ctrl+B — переключить видимость пароля\n",
	))

	return b.String()
}

// moveFocus переключает фокус вперёд (forward=true) или назад (forward=false)
func moveFocus(m Model, forward bool) Model {
	if forward {
		m.focusedInput++
		if m.focusedInput >= len(m.widgets) {
			m.focusedInput = 0
		}
	} else {
		m.focusedInput--
		if m.focusedInput < 0 {
			m.focusedInput = len(m.widgets) - 1
		}
	}
	return m
}

func focusField(m Model) Model {
	for i := range m.widgets {
		m.widgets[i].setFocus(i == m.focusedInput)
	}
	return m
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

	fields := m.ExtractFields()

	// 4) Валидация и обновление самой сущности (реализовано в доменной модели)
	if err := fe.UpdateFromFields(fields); err != nil {
		m.editErr = err
		return m
	}

	m.editEntity = fe // для чёткости присвоим
	m.editErr = nil
	return m
}
