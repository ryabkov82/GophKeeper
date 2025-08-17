package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

// formWidget представляет отдельное поле формы, которое может быть:
// обычным текстовым вводом (textinput) или многострочным textarea.
// Также поддерживает работу с маской ввода через maskedInput.
type formWidget struct {
	isTextarea  bool            // true, если виджет является textarea
	input       textinput.Model // текстовое поле ввода
	textarea    textarea.Model  // многострочное поле ввода
	field       forms.FormField // исходное описание поля формы
	maskedInput *MaskedInput    // объект для работы с маской, nil если маски нет
	fullscreen  bool            // поле редактируется в полноэкранном режиме
}

// setFocus устанавливает фокус на виджет или снимает его.
// При установке фокуса активируется соответствующий курсор и визуальный стиль.
func (w *formWidget) setFocus(focused bool) {
	if w.isTextarea {
		if focused {
			w.textarea.Focus()
		} else {
			w.textarea.Blur()
		}
	} else {
		if focused {
			w.input.Focus()
		} else {
			w.input.Blur()
		}
	}
}

// initFormInputsFromFields создаёт срез formWidget на основе массива FormField.
// Поддерживает следующие типы полей:
// - "multiline" — создаётся textarea.
// - "password" — создаётся текстовое поле с EchoPassword.
// - обычный ввод — создаётся textinput с ограничением длины из field.MaxLength.
// Если задано поле Mask, создаётся maskedInput и отображается сразу с подчеркиваниями.
func initFormInputsFromFields(fields []forms.FormField) []formWidget {
	widgets := make([]formWidget, len(fields))

	for i, field := range fields {
		w := formWidget{field: field}

		switch strings.ToLower(field.InputType) {
		case "multiline":
			ta := textarea.New()
			ta.Placeholder = ""
			ta.Cursor.Style = cursorStyle
			ta.SetValue(field.Value)
			ta.ShowLineNumbers = true
			ta.CharLimit = 0
			ta.Prompt = " "
			ta.SetWidth(100) // дефолтная ширина для обычных textarea
			w.isTextarea = true
			w.textarea = ta
			w.fullscreen = field.Fullscreen // сохраняем флаг

		default:
			ti := textinput.New()
			ti.Placeholder = ""
			ti.Prompt = " "
			ti.Cursor.Style = cursorStyle
			// если MaxLength задан — используем его, иначе дефолт 256
			if field.MaxLength > 0 {
				ti.CharLimit = field.MaxLength
			} else {
				ti.CharLimit = 256
			}
			ti.EchoMode = textinput.EchoNormal
			if strings.ToLower(field.InputType) == "password" {
				ti.EchoMode = textinput.EchoPassword
				ti.EchoCharacter = '•'
			}

			w.isTextarea = false
			w.input = ti

			// --- Работа с маской ---
			if field.Mask != "" {
				mi := NewMaskedInput(field.Mask, field.Value)
				w.maskedInput = &mi
				// сразу отображаем маску с подчеркиваниями
				w.input.SetValue(w.maskedInput.Display())
				w.input.SetCursor(w.maskedInput.CursorPos)
			} else {
				// обычное значение без маски
				w.input.SetValue(field.Value)
			}

		}

		w.setFocus(i == 0)
		widgets[i] = w
	}

	return widgets
}
