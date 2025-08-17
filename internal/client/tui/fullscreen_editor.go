package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func initFullscreenForm(m Model) Model {
	m.fullscreenErr = nil

	if m.fullscreenWidget == nil {
		m.fullscreenErr = fmt.Errorf("no widget selected for fullscreen edit")
		return m
	}

	w := m.fullscreenWidget
	if !w.isTextarea {
		m.fullscreenErr = fmt.Errorf("selected widget is not a textarea")
		return m
	}

	fullscreenTA := textarea.New()
	fullscreenTA.MaxHeight = 0               //снимаем ограничение на 100 строк
	fullscreenTA.SetWidth(m.termWidth - 2)   // оставляем по 1 символу слева/справа
	fullscreenTA.SetHeight(m.termHeight - 8) //
	fullscreenTA.SetValue(w.textarea.Value())
	for fullscreenTA.Line() > 0 {
		fullscreenTA.CursorUp()
	}
	fullscreenTA.Cursor.Style = cursorStyle
	fullscreenTA.Focus()

	// Обновляем fullscreenWidget в модели
	m.fullscreenWidget = &formWidget{
		field:      w.field,
		textarea:   fullscreenTA, // оригинальная textarea для хранения текста
		isTextarea: true,
		fullscreen: true,
	}

	return m
}

func updateFullscreenForm(m Model, msg tea.Msg) (Model, tea.Cmd) {

	if m.fullscreenWidget == nil {
		return m, nil
	}

	w := m.fullscreenWidget

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		step := m.termHeight - 8 // шаг прокрутки, учитывая заголовок и подсказку
		switch key {

		case "pgup":
			for i := 0; i < step; i++ {
				w.textarea.CursorUp()
			}
			w.textarea.CursorStart()
			m.fullscreenWidget = w
			return m, nil
		case "pgdown":
			for i := 0; i < step; i++ {
				w.textarea.CursorDown()
			}
			w.textarea.CursorEnd()
			m.fullscreenWidget = w
			return m, nil
		case "ctrl+home":
			for w.textarea.Line() > 0 {
				w.textarea.CursorUp()
			}
			w.textarea.CursorStart()

			m.fullscreenWidget = w
			return m, nil

		case "ctrl+end":

			// Бесконечный цикл, который прервётся когда достигнем последней строки
			for {
				currentLine := w.textarea.Line() // Запоминаем текущую строку
				w.textarea.CursorDown()          // Пробуем переместиться вниз
				w.textarea.CursorEnd()           // и в конец строки

				// Если после перемещения строка не изменилась - мы достигли конца
				if w.textarea.Line() == currentLine {
					break // Выходим из цикла
				}
			}
			m.fullscreenWidget = w
			return m, nil
		case "esc":
			// Отмена изменений, выход в обычную форму
			m.currentState = m.prevState // восстанавливаем предыдущий режим
			m.fullscreenWidget = nil
			m.fullscreenErr = nil
			return m, nil

		case "ctrl+s":
			// Сохраняем изменения в основной форме (UpdateFromFields)
			for i, f := range m.widgets {
				if f.field.Label == w.field.Label {
					f.textarea.SetValue(w.textarea.Value())
					m.widgets[i] = f
					break
				}
			}
			m.currentState = m.prevState // восстанавливаем предыдущий режим
			m.fullscreenWidget = nil
			return m, nil
		case "ctrl+v":
			clipText := clipboardRead()
			if clipText != "" {
				//newValue := w.textarea.Value() + clipText
				w.textarea.SetValue(clipText)
				m.fullscreenWidget = w
				return m, nil
			}
		case "ctrl+d": // Ctrl+D для очистки
			w.textarea.SetValue("")
			m.fullscreenWidget = w
			return m, nil
		case "ctrl+c":
			// Копируем всё содержимое textarea в буфер обмена
			if err := clipboard.WriteAll(w.textarea.Value()); err != nil {
				m.fullscreenErr = fmt.Errorf("ошибка копирования в буфер: %w", err)
			}
			return m, nil
		}
	}

	// Обновление
	var cmd tea.Cmd
	w.textarea, cmd = w.textarea.Update(msg)
	m.fullscreenWidget = w
	return m, cmd
}

func renderFullscreenForm(m Model) string {
	if m.fullscreenWidget == nil {
		return errorStyle.Render("Ошибка: нет выбранного поля для полноэкранного редактирования")
	}

	var b strings.Builder

	// Заголовок
	title := fmt.Sprintf("Полноэкранное редактирование: %s", m.fullscreenWidget.field.Label)
	b.WriteString(titleStyle.Render(title) + "\n\n")

	w := m.fullscreenWidget
	//vp := w.viewport

	// Обертка блока
	b.WriteString(formBlockStyle.Render(w.textarea.View()) + "\n")

	if m.fullscreenErr != nil {
		b.WriteString("\n" + errorStyle.Render("Ошибка: "+m.fullscreenErr.Error()))
	}

	// Подсказка по горячим клавишам
	b.WriteString("\n" + hintStyle.Render(
		"Esc: Выйти • Ctrl+S: Сохранить • Ctrl+C: копировать в буфер обмена • Ctrl+V: вставка из буфера обмена \n• Ctrl+D: очистка текста • Ctrl+Home: перемещение в начало • Ctrl+End: перемещение в конец\n",
	))

	return b.String()
}
