package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

// initFormInputsFromFields создает слайс formWidget на основе полей FormEntity.
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
			ta.ShowLineNumbers = false
			ta.CharLimit = 0
			ta.Prompt = " "

			w.isTextarea = true
			w.textarea = ta

		default:
			ti := textinput.New()
			ti.Placeholder = ""
			ti.Prompt = " "
			ti.Cursor.Style = cursorStyle
			ti.CharLimit = 256
			ti.EchoMode = textinput.EchoNormal
			if strings.ToLower(field.InputType) == "password" {
				ti.EchoMode = textinput.EchoPassword
				ti.EchoCharacter = '•'
			}

			w.isTextarea = false
			w.input = ti

			// --- Работа с маской ---
			if field.Mask != "" {
				w.maskedInput = NewMaskedInput(field.Mask, field.Value)
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
