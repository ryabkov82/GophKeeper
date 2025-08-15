package tui

import (
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// MaskedInput представляет поле ввода с маской.
// Позволяет ограничивать ввод определёнными символами в соответствии с маской.
// Маска содержит плейсхолдеры '#' и фиксированные символы.
// Raw хранит текущее состояние введённых символов, CursorPos — текущая позиция курсора.
type MaskedInput struct {
	Mask      string // маска поля, '#' обозначает placeholder
	Raw       []rune // текущее содержимое поля
	CursorPos int    // позиция курсора внутри поля
}

// NewMaskedInput создаёт новый MaskedInput с заданной маской и значением.
// Если в value есть символы, подходящие под маску, они вставляются на соответствующие позиции.
// Остальные позиции заполняются пробелами.
func NewMaskedInput(mask string, value string) MaskedInput {
	raw := make([]rune, len(mask))
	for i := range raw {
		raw[i] = ' ' // пустой placeholder
	}

	valRunes := []rune(value)
	valPos := 0

	for i, r := range mask {
		if isPlaceholder(r) {
			// вставляем символ из value, если он есть
			if valPos < len(valRunes) {
				// если текущий символ value подходит для этого placeholder
				if isRuneAllowedForMaskChar(r, valRunes[valPos]) {
					raw[i] = valRunes[valPos]
					valPos++
				} else {
					// если символ не подходит, пропускаем его
					valPos++
					i-- // остаёмся на этом placeholder
				}
			} else {
				raw[i] = ' ' // пустой
			}
		} else {
			// фиксированный символ маски
			raw[i] = r
			// если value совпадает с фиксированным символом, пропускаем его
			if valPos < len(valRunes) && valRunes[valPos] == r {
				valPos++
			}
		}
	}

	return MaskedInput{
		Mask:      mask,
		Raw:       raw,
		CursorPos: 0,
	}
}

// InsertRune вставляет один символ в текущую позицию курсора, если он соответствует маске.
func (m *MaskedInput) InsertRune(r rune) {
	for i := m.CursorPos; i < len(m.Mask); i++ {
		if isPlaceholder(rune(m.Mask[i])) {
			if isRuneAllowedForMaskChar(rune(m.Mask[i]), r) {
				m.Raw[i] = r
				m.CursorPos = i + 1
			}
			break
		}
	}
}

// InsertString вставляет строку символов последовательно, применяя InsertRune для каждого.
func (m *MaskedInput) InsertString(s string) {
	for _, r := range s {
		m.InsertRune(r)
	}
}

// Backspace удаляет символ перед курсором, если это placeholder, и перемещает курсор.
func (m *MaskedInput) Backspace() {
	for i := m.CursorPos - 1; i >= 0; i-- {
		if isPlaceholder(rune(m.Mask[i])) && m.Raw[i] != ' ' {
			m.Raw[i] = ' '
			m.CursorPos = i
			break
		}
	}
}

// Delete удаляет символ под курсором, если это placeholder.
func (m *MaskedInput) Delete() {
	for i := m.CursorPos; i < len(m.Mask) && i < len(m.Raw); i++ {
		if isPlaceholder(rune(m.Mask[i])) {
			// Если символ под курсором не пустой — очищаем
			if m.Raw[i] != ' ' {
				m.Raw[i] = ' '
			}
			break
		}
	}
}

// Home перемещает курсор на первый доступный placeholder.
func (m *MaskedInput) Home() {
	for i, r := range m.Mask {
		if isPlaceholder(r) {
			m.CursorPos = i
			break
		}
	}
}

// End перемещает курсор на последний введённый символ в маске.
func (m *MaskedInput) End() {
	last := 0
	for i, r := range m.Mask {
		if isPlaceholder(r) && m.Raw[i] != ' ' {
			last = i
		}
	}
	m.CursorPos = last
}

// Display возвращает строку для отображения в интерфейсе,
// заменяя пустые placeholder'ы на '_'.
func (m *MaskedInput) Display() string {
	out := make([]rune, len(m.Mask))
	for i, r := range m.Mask {
		if isPlaceholder(r) {
			if m.Raw[i] == ' ' {
				out[i] = '_' // отображаем пустой символ как подчеркивание
			} else {
				out[i] = m.Raw[i]
			}
		} else {
			out[i] = r
		}
	}
	return string(out)
}

// MoveLeft перемещает курсор на предыдущий placeholder слева.
func (m *MaskedInput) MoveLeft() {
	for i := m.CursorPos - 1; i >= 0; i-- {
		if isPlaceholder(rune(m.Mask[i])) {
			m.CursorPos = i
			break
		}
	}
}

// MoveRight перемещает курсор на следующий placeholder справа.
func (m *MaskedInput) MoveRight() {
	for i := m.CursorPos + 1; i < len(m.Mask); i++ {
		if isPlaceholder(rune(m.Mask[i])) {
			m.CursorPos = i
			break
		}
	}
}

// HandleMaskedInput обновляет maskedInput и соответствующий textinput.Model
// на основе нажатой клавиши (backspace, delete, home, end, left, right, ctrl+v) или введённого символа.
func HandleMaskedInput(w *formWidget, key string, msg tea.KeyMsg) {
	mi := w.maskedInput
	switch key {
	case "backspace":
		mi.Backspace()
	case "delete":
		mi.Delete()
	case "home":
		mi.Home()
	case "end":
		mi.End()
	case "ctrl+v":
		mi.InsertString(clipboardRead())
	case "left":
		mi.MoveLeft()
	case "right":
		mi.MoveRight()
	default:
		if len(msg.Runes) > 0 {
			mi.InsertRune(msg.Runes[0])
		}
	}

	w.input.SetValue(mi.Display())
	w.input.SetCursor(mi.CursorPos)
}

// isPlaceholder проверяет, является ли символ маски плейсхолдером (#).
func isPlaceholder(r rune) bool {
	return r == '#'
}

// isRuneAllowedForMaskChar проверяет, разрешён ли вводимый символ input для данного символа маски maskChar.
func isRuneAllowedForMaskChar(maskChar, input rune) bool {
	switch maskChar {
	case '#':
		return input >= '0' && input <= '9'
	default:
		return false
	}
}

// clipboardRead возвращает текст из буфера обмена.
// В случае ошибки возвращает пустую строку.
func clipboardRead() string {
	text, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	return text
}
