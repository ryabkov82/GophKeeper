package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMaskedInput_InsertRuneAndDisplay(t *testing.T) {
	mi := NewMaskedInput("####-####", "")
	mi.InsertRune('1')
	mi.InsertRune('2')
	mi.InsertRune('3')
	mi.InsertRune('4')

	expectedRaw := []rune{'1', '2', '3', '4', '-', ' ', ' ', ' ', ' '}
	if string(mi.Raw) != string(expectedRaw) {
		t.Errorf("expected Raw=%v, got %v", expectedRaw, mi.Raw)
	}

	expectedDisplay := "1234-____"
	if mi.Display() != expectedDisplay {
		t.Errorf("expected Display=%s, got %s", expectedDisplay, mi.Display())
	}
}

func TestMaskedInput_InsertString(t *testing.T) {
	mi := NewMaskedInput("####-####", "")
	mi.InsertString("5678")
	if mi.Display() != "5678-____" {
		t.Errorf("expected Display='5678-____', got %s", mi.Display())
	}
}

func TestMaskedInput_Backspace(t *testing.T) {
	mi := NewMaskedInput("####", "1234")
	mi.CursorPos = 4
	mi.Backspace()
	if mi.Raw[3] != ' ' || mi.CursorPos != 3 {
		t.Errorf("expected Raw[3]=' ' and CursorPos=3, got Raw=%v, CursorPos=%d", mi.Raw, mi.CursorPos)
	}
}

func TestMaskedInput_Delete(t *testing.T) {
	mi := NewMaskedInput("####", "1234")
	mi.CursorPos = 1
	mi.Delete()
	if mi.Raw[1] != ' ' {
		t.Errorf("expected Raw[1]=' ', got %v", mi.Raw)
	}
}

func TestMaskedInput_HomeEnd(t *testing.T) {
	mi := NewMaskedInput("####-####", "12")
	mi.Home()
	if mi.CursorPos != 0 {
		t.Errorf("expected CursorPos=0, got %d", mi.CursorPos)
	}

	mi.End()
	if mi.CursorPos != 1 {
		t.Errorf("expected CursorPos=1, got %d", mi.CursorPos)
	}
}

func TestMaskedInput_MoveLeftRight(t *testing.T) {
	mi := NewMaskedInput("##-##", "12")
	mi.CursorPos = 2
	mi.MoveLeft()
	if mi.CursorPos != 1 {
		t.Errorf("expected CursorPos=1, got %d", mi.CursorPos)
	}

	mi.MoveRight()
	if mi.CursorPos != 3 {
		t.Errorf("expected CursorPos=3, got %d", mi.CursorPos)
	}
}

func TestMaskedInput_MaskWithFixedChars(t *testing.T) {
	mi := NewMaskedInput("+7 (###) ###-##-##", "9123456789")
	display := mi.Display()
	expected := "+7 (912) 345-67-89"
	if display != expected {
		t.Errorf("expected Display=%s, got %s", expected, display)
	}
}

func TestHandleMaskedInput_InsertRune(t *testing.T) {
	mi := NewMaskedInput("####", "")
	w := &formWidget{
		maskedInput: &mi,
	}
	msg := tea.KeyMsg{Runes: []rune{'5'}}
	HandleMaskedInput(w, "5", msg)
	if w.maskedInput.Raw[0] != '5' {
		t.Errorf("expected first rune '5', got %c", w.maskedInput.Raw[0])
	}
}
