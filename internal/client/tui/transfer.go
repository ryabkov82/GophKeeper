package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// ----- режимы -----
type transferMode int

const (
	modeUpload transferMode = iota
	modeDownload
)

// ----- сообщения -----
type transferProgressMsg struct{ sent int64 }
type transferDoneMsg struct{}
type transferErrorMsg struct{ err error }

// ----- внутренний VM экрана передачи -----
type transferVM struct {
	// deps
	svc    contracts.BinaryTransferCapable
	data   *model.BinaryData // мета (для upload и имени/размера при download)
	dataID string            // ID для download (если data == nil)

	// ui
	mode  transferMode
	input textinput.Model

	// прогресс
	inFlight bool
	err      error
	total    int64
	sent     int64

	// скорость/ETA
	startedAt    time.Time
	lastTickAt   time.Time
	lastTickSent int64
	speedBps     float64

	// async
	ctx    context.Context
	cancel context.CancelFunc
	progCh chan int64
	doneCh chan error
}

// ===== init/update/render в стиле проекта =====

// initTransferForm — инициализация формы Upload/Download
func initTransferForm(m Model, mode transferMode, data *model.BinaryData, dataID string) Model {
	m.prevState = m.currentState
	m.currentState = "file_transfer"

	// достаём BinaryDataService
	var svc contracts.BinaryTransferCapable
	if s, ok := m.services[contracts.TypeFiles].(contracts.BinaryTransferCapable); ok {
		svc = s
	}

	ti := newInputField("")
	ti.Prompt = " Путь: "
	if mode == modeUpload {
		ti.Placeholder = "Источник (локальный файл)"
	} else {
		ti.Placeholder = "Назначение (куда сохранить)"
	}
	ti.Focus()

	// запретить прямой старт в Download, если нет ClientPath
	if mode == modeDownload && !canSwitchToDownload(data) {
		mode = modeUpload // мягко переключаем на Upload
		// и сохраним причину, чтобы показать её пользователю
		m.transfer.err = fmt.Errorf("скачивание недоступно: файл ещё не загружался")
	}

	m.transfer = transferVM{
		svc:    svc,
		data:   data,
		dataID: dataID,

		mode:  mode,
		input: ti,

		inFlight: false,
		err:      nil,
		total:    0,
		sent:     0,
	}

	return m
}

// updateTransfer — обработчик событий формы передачи
func updateTransfer(m Model, msg tea.Msg) (Model, tea.Cmd) {
	t := m.transfer

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+u":
			if !t.inFlight {
				t.mode = modeUpload
				t.input.Placeholder = "Источник (локальный файл)"
			}
		case "ctrl+d":
			if !t.inFlight {
				if canSwitchToDownload(t.data) {
					t.mode = modeDownload
					t.input.Placeholder = "Назначение (куда сохранить)"
				} else {
					t.err = fmt.Errorf("скачивание недоступно: файл ещё не загружался")
				}
			}
		case "enter":
			if !t.inFlight {
				return startTransfer(m)
			}
		case "esc", "ctrl+c":
			// отмена/назад
			if t.inFlight && t.cancel != nil {
				t.cancel()
			}
			m.currentState = m.prevState
			if t.data.ID != "" {
				return loadAndShowItem(m, t.data.ID)
			}
			return m, nil
		}

	case transferProgressMsg:
		// апдейтим количество и скорость
		now := time.Now()
		dt := now.Sub(t.lastTickAt).Seconds()
		if dt > 0 {
			t.speedBps = float64(msg.sent-t.lastTickSent) / dt
		}
		t.lastTickAt = now
		t.lastTickSent = msg.sent

		t.sent = msg.sent
		if t.total > 0 && t.sent > t.total {
			t.sent = t.total
		}
		m.transfer = t
		return m, listenTransferProgress(t.progCh)

	case transferDoneMsg:
		t.inFlight = false
		t.err = nil
		m.transfer = t
		return m, nil

	case transferErrorMsg:
		t.inFlight = false
		t.err = msg.err
		m.transfer = t
		return m, nil
	}

	// делегируем ввод инпуту, если не идёт передача
	if !t.inFlight {
		var cmd tea.Cmd
		t.input, cmd = t.input.Update(msg)
		m.transfer = t
		return m, cmd
	}

	m.transfer = t
	return m, nil
}

// renderTransfer — рендер в стиле остальных форм
func renderTransfer(m Model) string {
	t := m.transfer
	var b strings.Builder

	title := "Загрузка файла на сервер"
	if t.mode == modeDownload {
		title = "Скачивание файла с сервера"
	}
	b.WriteString(titleStyle.Render(title) + "\n\n")
	b.WriteString(t.input.View() + "\n\n")

	if t.inFlight {
		// проценты/бар
		if t.total > 0 {
			r := float64(t.sent) / float64(t.total)
			if r < 0 {
				r = 0
			}
			if r > 1 {
				r = 1
			}
			b.WriteString(renderAsciiBar(r, max(20, m.termWidth-20)) + "\n")
			b.WriteString(fmt.Sprintf("  %s / %s (%.1f%%)",
				humanBytes(t.sent), humanBytes(t.total), r*100))
		} else {
			b.WriteString(renderAsciiBar(0, max(20, m.termWidth-20)) + "\n")
			b.WriteString(fmt.Sprintf("  Передано: %s (размер неизвестен)", humanBytes(t.sent)))
		}

		/*
			// скорость/ETA
			if t.speedBps > 0 && t.total > 0 && t.sent <= t.total {
				remain := float64(t.total - t.sent)
				eta := time.Duration(remain/t.speedBps) * time.Second
				b.WriteString(fmt.Sprintf(" • ~%s/s • ETA %s", humanBytes(int64(t.speedBps)), eta.Truncate(time.Second)))
			} else if t.speedBps > 0 {
				b.WriteString(fmt.Sprintf(" • ~%s/s", humanBytes(int64(t.speedBps))))
			}
		*/

		b.WriteString("\n\n" + hintStyle.Render("Esc: отменить/назад • Ctrl+U/Ctrl+D: режим Upload/Download"))
	} else {
		b.WriteString(hintStyle.Render("Enter: старт • Esc: назад • Ctrl+U/Ctrl+D: режим Upload/Download"))
		if t.err != nil {
			b.WriteString("\n\n" + errorStyle.Render("Ошибка: "+t.err.Error()))
		}
	}
	return b.String()
}

// ===== команды/утилиты =====

func startTransfer(m Model) (Model, tea.Cmd) {

	t := m.transfer

	// на всякий случай отсекаем недопустимый Download
	if t.mode == modeDownload && !canSwitchToDownload(t.data) {
		t.err = fmt.Errorf("скачивание недоступно: файл ещё не загружался")
		m.transfer = t
		return m, nil
	}

	path := strings.TrimSpace(t.input.Value())
	if path == "" {
		t.err = fmt.Errorf("укажите путь")
		m.transfer = t
		return m, nil
	}
	if t.svc == nil {
		t.err = fmt.Errorf("BinaryDataService недоступен")
		m.transfer = t
		return m, nil
	}

	// total и нормализация пути
	if t.mode == modeUpload {
		st, err := os.Stat(path)
		if err != nil {
			t.err = err
			m.transfer = t
			return m, nil
		}
		if st.IsDir() {
			t.err = fmt.Errorf("ожидается файл, а не каталог")
			m.transfer = t
			return m, nil
		}
		t.total = st.Size()
	} else {
		// каталог -> доклеим имя
		if fi, err := os.Stat(path); err == nil && fi.IsDir() {
			name := "download.bin"
			path = filepath.Join(path, name)
		}
		if t.data != nil && t.data.Size > 0 {
			t.total = t.data.Size
		} else {
			t.total = 0
		}
	}

	// запуск фоновой операции
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.progCh = make(chan int64, 32)
	t.doneCh = make(chan error, 1)
	t.inFlight = true
	t.err = nil
	t.sent = 0
	t.startedAt = time.Now()
	t.lastTickAt = t.startedAt
	t.lastTickSent = 0
	t.speedBps = 0

	go func(tt *transferVM, p string) {
		var err error
		if tt.mode == modeUpload {
			err = tt.svc.UploadBinaryData(tt.ctx, tt.data, p, tt.progCh)
		} else {
			id := tt.dataID
			if id == "" && tt.data != nil {
				id = tt.data.ID
			}
			err = tt.svc.DownloadBinaryData(tt.ctx, id, p, tt.progCh)
		}
		close(tt.progCh)
		tt.doneCh <- err
	}(&t, path)

	m.transfer = t
	return m, tea.Batch(listenTransferProgress(t.progCh), waitTransferDone(t.doneCh))
}

func listenTransferProgress(ch <-chan int64) tea.Cmd {
	return func() tea.Msg {
		if n, ok := <-ch; ok {
			return transferProgressMsg{sent: n}
		}
		return nil
	}
}

func waitTransferDone(done <-chan error) tea.Cmd {
	return func() tea.Msg {
		if err := <-done; err != nil {
			return transferErrorMsg{err: err}
		}
		return transferDoneMsg{}
	}
}

func renderAsciiBar(ratio float64, width int) string {
	if width < 10 {
		width = 10
	}
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("·", width-filled) + "]"
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// upload_download.go — рядом с остальными хелперами
func canSwitchToDownload(data *model.BinaryData) bool {
	if data == nil {
		return false
	}
	return strings.TrimSpace(data.ClientPath) != ""
}
