package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

func initEditForm(m Model) Model {

	m.editErr = nil
	if fe, err := forms.Adapt(m.editEntity); err == nil {
		formFields := fe.FormFields()
		m.widgets, m.focusedInput = initFormInputsFromFields(formFields, m.termWidth)
	} else {
		m.editErr = err
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
			return handleListSelection(m, m.currentType)
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
		case "f2":
			if len(m.widgets) == 0 {
				return m, nil
			}
			w := m.widgets[m.focusedInput]
			if w.isTextarea && w.fullscreen {
				// Переключаемся в отдельный режим редактирования fullscreen
				m.prevState = m.currentState // сохраняем текущее состояние
				m.currentState = "fullscreen_edit"
				m.fullscreenWidget = &w // сохраняем текущий виджет
				m = initFullscreenForm(m)
			}
			return m, nil

		case "ctrl+u":
			if m.currentType == contracts.TypeFiles {
				m = updateEditEntityFromInputs(m)
				var bd *model.BinaryData
				if v, ok := m.editEntity.(*model.BinaryData); ok {
					bd = v
				}
				m = initTransferForm(m, modeUpload, bd, "")
			}
			return m, nil

		case "ctrl+d":
			if m.currentType == contracts.TypeFiles {
				var (
					bd *model.BinaryData
					id string
				)
				if v, ok := m.editEntity.(*model.BinaryData); ok {
					bd = v
					id = v.ID
				}

				// запрет, если клиент ещё ничего не загружал (ClientPath пуст)
				if !canSwitchToDownload(bd) {
					m.editErr = fmt.Errorf("скачивание недоступно: файл ещё не загружался")
					return m, nil
				}

				// запрет, если у записи нет ID — нечего скачивать
				if strings.TrimSpace(id) == "" {
					m.editErr = fmt.Errorf("скачивание недоступно: сначала сохраните запись")
					return m, nil
				}

				m = initTransferForm(m, modeDownload, bd, id)
			}
			return m, nil
		}

		if len(m.widgets) == 0 {
			return m, nil
		}

		w := &m.widgets[m.focusedInput]
		if w.readonly {
			return m, nil
		}
		if w.fullscreen {
			// не обрабатываем никакой ввод, просто возвращаем модель без изменений
			return m, nil
		}

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

	for i, widget := range m.widgets {
		label := widget.field.Label + ": "
		if widget.readonly {
			label = readonlyStyle.Render(label)
		} else if i == m.focusedInput {
			label = activeFieldStyle.Render(label)
		} else {
			label = inactiveFieldStyle.Render(label)
		}

		if widget.isTextarea {
			if widget.fullscreen {
				label += " (F2 для полноэкранного просмотра/редактирования)"
				/*
					content := widget.textarea.Value() // получаем всё содержимое
					lines := strings.Split(content, "\n")
					if len(lines) > 5 {
						lines = lines[:5]
						lines = append(lines, "…") // добавляем многоточие
					}
					content = strings.Join(lines, "\n")
					// делаем readonly-стиль
					b.WriteString(label)
					if len(content) > 0 {
						content = readonlyStyle.Render(content)
						b.WriteString(formBlockStyle.Render(content) + "\n")
					}
				*/
			}
			//b.WriteString(label + "\n" + formBlockStyle.Render(widget.textarea.View()) + "\n")
			blockStyle := formBlockStyle
			if widget.readonly {
				blockStyle = blockStyle.Foreground(lipgloss.Color("244"))
			}
			b.WriteString(label + "\n" + blockStyle.Render(widget.textarea.View()) + "\n")
		} else {
			//b.WriteString(label + widget.input.View() + "\n")
			inputView := widget.input.View()
			if widget.readonly {
				inputView = readonlyStyle.Render(inputView)
			}
			b.WriteString(label + inputView + "\n")
		}
	}

	if m.editErr != nil {
		b.WriteString("\n" + errorStyle.Render("Ошибка: "+m.editErr.Error()))
	}

	hint := "Esc: Назад • Ctrl+S: Сохранить • Tab: Следующее поле "
	switch m.currentType {
	case contracts.TypeFiles:
		hint += "• Ctrl+U: загрузить файл • Ctrl+D: скачать файл\n"
	default:
		hint += "• Ctrl+B: переключить видимость пароля\n"
	}

	b.WriteString("\n" + hintStyle.Render(hint))

	return b.String()
}

// moveFocus переключает фокус вперёд (forward=true) или назад (forward=false)
func moveFocus(m Model, forward bool) Model {

	if len(m.widgets) == 0 {
		return m
	}
	start := m.focusedInput
	for {
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
		if !m.widgets[m.focusedInput].readonly {
			break
		}
		if m.focusedInput == start {
			break
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
	} else {
		if idGetter, ok := m.editEntity.(forms.Identifiable); ok {
			return loadAndShowItem(m, idGetter.GetID())
		}
	}

	return handleListSelection(m, m.currentType)
}

// updateEditEntityFromInputs обновляет m.editEntity значениями из m.inputs.
// Возвращает обновлённую модель (m.editErr заполняется при ошибке).
func updateEditEntityFromInputs(m Model) Model {

	fe, err := forms.Adapt(m.editEntity)
	if err != nil {
		m.editErr = err
		return m
	}

	fields := m.ExtractFields()

	// 4) Валидация и обновление самой сущности (реализовано в доменной модели)
	if err := fe.UpdateFromFields(fields); err != nil {
		m.editErr = err
		return m
	}

	m.editErr = nil
	return m
}
