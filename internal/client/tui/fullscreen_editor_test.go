package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/stretchr/testify/assert"
)

// helper: создать минимальную модель для fullscreen-режима
func makeFullscreenModel(initial string) Model {
	// подготовим виджет с textarea
	w := formWidget{
		field:      forms.FormField{Label: "Content", InputType: "multiline"},
		textarea:   textarea.New(),
		isTextarea: true,
		fullscreen: true,
	}
	w.textarea.SetValue(initial)

	m := Model{
		termWidth:        100,
		termHeight:       30,
		currentState:     "fullscreen_edit",
		prevState:        "edit",
		fullscreenWidget: &w,
		// widgets заполним при необходимости (для Ctrl+S)
	}
	// проинициализируем полноэкранный редактор
	m = initFullscreenForm(m)
	return m
}

func TestInitFullscreenForm_OK(t *testing.T) {
	m := makeFullscreenModel("hello")
	assert.Nil(t, m.fullscreenErr)
	assert.NotNil(t, m.fullscreenWidget)
	assert.True(t, m.fullscreenWidget.isTextarea)
	assert.True(t, m.fullscreenWidget.fullscreen)
	assert.Equal(t, "hello", m.fullscreenWidget.textarea.Value())
	// важные настройки
	assert.Equal(t, 0, m.fullscreenWidget.textarea.MaxHeight)
}

func TestInitFullscreenForm_NoWidget(t *testing.T) {
	m := Model{termWidth: 80, termHeight: 24}
	m = initFullscreenForm(m)
	assert.Error(t, m.fullscreenErr)
	assert.Contains(t, m.fullscreenErr.Error(), "no widget")
}

func TestInitFullscreenForm_NotTextarea(t *testing.T) {
	w := formWidget{
		field:      forms.FormField{Label: "Title", InputType: "text"},
		isTextarea: false,
	}
	m := Model{
		termWidth:        80,
		termHeight:       24,
		fullscreenWidget: &w,
	}
	m = initFullscreenForm(m)
	assert.Error(t, m.fullscreenErr)
	assert.Contains(t, m.fullscreenErr.Error(), "not a textarea")
}

func TestUpdateFullscreenForm_EnterInsertsNewline(t *testing.T) {
	m := makeFullscreenModel("h")
	m2, _ := updateFullscreenForm(m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "\nh", m2.fullscreenWidget.textarea.Value())
}

func TestUpdateFullscreenForm_CtrlHomeAndCtrlEnd(t *testing.T) {
	m := makeFullscreenModel("hello\nworld")

	// Ctrl+Home, затем ввести "X" → должно оказаться в самом начале
	m1, _ := updateFullscreenForm(m, tea.KeyMsg{Type: tea.KeyCtrlHome})
	m1, _ = updateFullscreenForm(m1, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("X")})
	assert.True(t, strings.HasPrefix(m1.fullscreenWidget.textarea.Value(), "X"))

	// Ctrl+End, затем ввести "Y" → в самом конце
	m2, _ := updateFullscreenForm(m1, tea.KeyMsg{Type: tea.KeyCtrlEnd})
	m2, _ = updateFullscreenForm(m2, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")})
	assert.True(t, strings.HasSuffix(m2.fullscreenWidget.textarea.Value(), "Y"))
}

func TestUpdateFullscreenForm_PgUpPgDn_DoNotChangeText(t *testing.T) {
	orig := strings.Repeat("line\n", 200)
	m := makeFullscreenModel(orig)

	// PgUp/PgDn должны менять позицию курсора/скролла, но не текст
	m1, _ := updateFullscreenForm(m, tea.KeyMsg{Type: tea.KeyPgUp})
	assert.Equal(t, orig, m1.fullscreenWidget.textarea.Value())

	m2, _ := updateFullscreenForm(m1, tea.KeyMsg{Type: tea.KeyPgDown})
	assert.Equal(t, orig, m2.fullscreenWidget.textarea.Value())
}

func TestUpdateFullscreenForm_CtrlD_Clears(t *testing.T) {
	m := makeFullscreenModel("some text")
	m2, _ := updateFullscreenForm(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	assert.Equal(t, "", m2.fullscreenWidget.textarea.Value())
}

func TestUpdateFullscreenForm_Esc_ExitsAndRestoresState(t *testing.T) {
	m := makeFullscreenModel("text")
	m.prevState = "edit"
	m.currentState = "fullscreen_edit"

	m2, _ := updateFullscreenForm(m, tea.KeyMsg{Type: tea.KeyEsc})
	assert.Equal(t, "edit", m2.currentState)
	assert.Nil(t, m2.fullscreenWidget)
	assert.Nil(t, m2.fullscreenErr)
}

func TestUpdateFullscreenForm_CtrlS_WritesBackToWidgets(t *testing.T) {
	// в общей форме есть поле с таким же Label
	orig := "alpha"
	m := makeFullscreenModel(orig)

	// Подготовим список виджетов формы с таким же ярлыком
	w := formWidget{
		field:      forms.FormField{Label: "Content", InputType: "multiline"},
		textarea:   textarea.New(),
		isTextarea: true,
	}
	w.textarea.SetValue("OLD")
	m.widgets = []formWidget{w}

	// Изменим текст в fullscreen, затем Ctrl+S → в widgets должен обновиться
	m.fullscreenWidget.textarea.SetValue("NEW VALUE")
	m2, _ := updateFullscreenForm(m, tea.KeyMsg{Type: tea.KeyCtrlS})

	assert.Equal(t, "edit", m2.currentState)
	assert.Nil(t, m2.fullscreenWidget)
	assert.Equal(t, "NEW VALUE", m2.widgets[0].textarea.Value())
}

func TestRenderFullscreenForm_NoWidget(t *testing.T) {
	m := Model{}
	out := renderFullscreenForm(m)
	assert.Contains(t, out, "Ошибка")
}

func TestRenderFullscreenForm_WithWidget(t *testing.T) {
	ta := textarea.New()
	ta.SetValue("hello")
	w := &formWidget{isTextarea: true, textarea: ta, field: forms.FormField{Label: "f1"}, fullscreen: true}
	m := Model{fullscreenWidget: w}

	out := renderFullscreenForm(m)
	assert.Contains(t, out, "Полноэкранное редактирование")
	assert.Contains(t, out, "hello")
}
