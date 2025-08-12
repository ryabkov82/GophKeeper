package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

// initFormInputsFromFields создает слайс textinput.Model на основе полей FormEntity.
func initFormInputsFromFields(entity forms.FormEntity) []textinput.Model {
	fields := entity.FormFields()
	inputs := make([]textinput.Model, len(fields))

	for i, field := range fields {
		ti := textinput.New()
		ti.Placeholder = ""
		ti.Prompt = " "
		ti.Cursor.Style = lipgloss.NewStyle().Background(lipgloss.Color("15"))
		ti.SetValue(field.Value)
		ti.CharLimit = 256 // можно настроить лимит символов

		// По умолчанию текст виден
		ti.EchoMode = textinput.EchoNormal

		// Если поле - пароль, скрываем ввод
		if strings.ToLower(field.InputType) == "password" {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}

		// Можно расширить для multiline или других типов, например:
		// if field.InputType == "multiline" { ... }

		// Фокусируем первый элемент, остальные нет
		if i == 0 {
			ti.Focus()
		} else {
			ti.Blur()
		}

		inputs[i] = ti
	}

	return inputs
}

// extractFieldsFromInputs получает заполненные значения из textinput.Model
// и возвращает слайс FormField с актуальными значениями.
// Порядок должен совпадать с исходным (FormFields).
func extractFieldsFromInputs(fields []forms.FormField, inputs []textinput.Model) []forms.FormField {
	// берём min длину, если размеры не совпадают — обновляем по минимуму
	n := len(fields)
	if len(inputs) < n {
		n = len(inputs)
	}
	for i := 0; i < n; i++ {
		fields[i].Value = inputs[i].Value()
	}
	return fields
}
