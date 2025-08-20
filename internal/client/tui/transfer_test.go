package tui

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBinaryTransferService implements contracts.BinaryTransferCapable for tests
// but keep in same package; not imported contract interface.
type mockBinaryTransferService struct {
	uploadCalled   bool
	downloadCalled bool
	uploadErr      error
	downloadErr    error
}

func (f *mockBinaryTransferService) List(ctx context.Context) ([]contracts.ListItem, error) {
	return nil, nil
}
func (f *mockBinaryTransferService) Get(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}
func (f *mockBinaryTransferService) Create(ctx context.Context, v interface{}) error {
	return nil
}
func (f *mockBinaryTransferService) Update(ctx context.Context, id string, v interface{}) error {
	return nil
}
func (f *mockBinaryTransferService) Delete(ctx context.Context, id string) error { return nil }

func (m *mockBinaryTransferService) UploadBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progress chan<- int64) error {
	m.uploadCalled = true
	// send some progress
	progress <- 1
	return m.uploadErr
}

func (m *mockBinaryTransferService) DownloadBinaryData(ctx context.Context, dataID, destPath string, progress chan<- int64) error {
	m.downloadCalled = true
	progress <- 1
	return m.downloadErr
}

func TestCanSwitchToDownload(t *testing.T) {
	tests := []struct {
		name string
		data *model.BinaryData
		want bool
	}{
		{"nil data", nil, false},
		{"empty path", &model.BinaryData{ClientPath: ""}, false},
		{"with path", &model.BinaryData{ClientPath: "/tmp/file"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, canSwitchToDownload(tc.data))
		})
	}
}

func TestHumanBytes(t *testing.T) {
	assert.Equal(t, "0 B", humanBytes(0))
	assert.Equal(t, "1023 B", humanBytes(1023))
	assert.Equal(t, "1.0 KiB", humanBytes(1024))
	assert.Equal(t, "5.0 MiB", humanBytes(5*1024*1024))
}

func TestListenTransferProgress(t *testing.T) {
	ch := make(chan int64, 1)
	ch <- 42
	cmd := listenTransferProgressThrottle(ch, 1)
	msg := cmd()
	progress, ok := msg.(transferProgressMsg)
	require.True(t, ok)
	assert.Equal(t, int64(42), progress.sent)

	emptyCh := make(chan int64)
	close(emptyCh)
	cmd = listenTransferProgressThrottle(emptyCh, 1)
	assert.Nil(t, cmd())
}

func TestWaitTransferDone(t *testing.T) {
	done := make(chan error, 1)
	done <- nil
	cmd := waitTransferDone(done)
	msg := cmd()
	assert.IsType(t, transferDoneMsg{}, msg)

	errCh := make(chan error, 1)
	errCh <- errors.New("boom")
	cmd = waitTransferDone(errCh)
	msg = cmd()
	errMsg, ok := msg.(transferErrorMsg)
	require.True(t, ok)
	assert.EqualError(t, errMsg.err, "boom")
}

func TestStartTransfer_EmptyPath(t *testing.T) {
	ti := newInputField("")
	ti.SetValue("")
	m := Model{transfer: transferVM{mode: modeUpload, input: ti}}
	m, cmd := startTransfer(m)
	assert.Nil(t, cmd)
	require.Error(t, m.transfer.err)
	assert.Equal(t, "укажите путь", m.transfer.err.Error())
}

func TestStartTransfer_NoService(t *testing.T) {
	ti := newInputField("")
	ti.SetValue("somepath")
	m := Model{transfer: transferVM{mode: modeUpload, input: ti}}
	m, cmd := startTransfer(m)
	assert.Nil(t, cmd)
	require.Error(t, m.transfer.err)
	assert.Equal(t, "BinaryDataService недоступен", m.transfer.err.Error())
}

func TestStartTransfer_UploadFile(t *testing.T) {
	f, err := os.CreateTemp("", "upload")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.Write([]byte("hello"))
	require.NoError(t, err)
	require.NoError(t, f.Close())

	svc := &mockBinaryTransferService{}
	ti := newInputField("")
	ti.SetValue(f.Name())
	m := Model{transfer: transferVM{mode: modeUpload, input: ti, svc: svc, data: &model.BinaryData{}}}

	m, cmd := startTransfer(m)
	require.NotNil(t, cmd)
	assert.True(t, m.transfer.inFlight)
	assert.Equal(t, int64(5), m.transfer.total)

	// allow goroutine to run
	time.Sleep(10 * time.Millisecond)
	assert.True(t, svc.uploadCalled)
}

func TestStartTransfer_DownloadNoClientPath(t *testing.T) {
	ti := newInputField("")
	ti.SetValue("/tmp/out")
	svc := &mockBinaryTransferService{}
	data := &model.BinaryData{ClientPath: ""}
	m := Model{transfer: transferVM{mode: modeDownload, input: ti, svc: svc, data: data}}
	m, cmd := startTransfer(m)
	assert.Nil(t, cmd)
	require.Error(t, m.transfer.err)
	assert.Contains(t, m.transfer.err.Error(), "скачивание недоступно")
}

func TestStartTransfer_UploadDir(t *testing.T) {
	dir := t.TempDir()
	ti := newInputField("")
	ti.SetValue(dir)
	svc := &mockBinaryTransferService{}
	m := Model{transfer: transferVM{mode: modeUpload, input: ti, svc: svc}}
	m, cmd := startTransfer(m)
	assert.Nil(t, cmd)
	require.Error(t, m.transfer.err)
	assert.Equal(t, "ожидается файл, а не каталог", m.transfer.err.Error())
}

func TestStartTransfer_CommandMessages(t *testing.T) {
	f, err := os.CreateTemp("", "upload")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	svc := &mockBinaryTransferService{}
	ti := newInputField("")
	ti.SetValue(f.Name())
	m := Model{transfer: transferVM{mode: modeUpload, input: ti, svc: svc, data: &model.BinaryData{}}}
	_, cmd := startTransfer(m)
	require.NotNil(t, cmd)
	time.Sleep(10 * time.Millisecond)
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	require.True(t, ok)
	var msgs []tea.Msg
	for _, c := range batch {
		msgs = append(msgs, c())
	}
	require.Len(t, msgs, 2)
	assert.IsType(t, transferProgressMsg{}, msgs[0])
	assert.IsType(t, transferDoneMsg{}, msgs[1])
}

func TestInitTransferForm_Upload(t *testing.T) {
	svc := &mockBinaryTransferService{}
	m := Model{
		currentState: "list",
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeFiles: svc,
		},
	}
	data := &model.BinaryData{ID: "1"}
	m2 := initTransferForm(m, modeUpload, data, "id1")
	assert.Equal(t, "list", m2.prevState)
	assert.Equal(t, "file_transfer", m2.currentState)
	assert.Equal(t, modeUpload, m2.transfer.mode)
	assert.Equal(t, svc, m2.transfer.svc)
	assert.Equal(t, data, m2.transfer.data)
	assert.Equal(t, "id1", m2.transfer.dataID)
	assert.Equal(t, "Источник (локальный файл)", m2.transfer.input.Placeholder)
	assert.Nil(t, m2.transfer.err)
}

func TestInitTransferForm_DownloadAllowed(t *testing.T) {
	svc := &mockBinaryTransferService{}
	m := Model{
		currentState: "list",
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeFiles: svc,
		},
	}
	data := &model.BinaryData{ClientPath: "/tmp/file"}
	m2 := initTransferForm(m, modeDownload, data, "id1")
	assert.Equal(t, modeDownload, m2.transfer.mode)
	assert.Equal(t, "Назначение (куда сохранить)", m2.transfer.input.Placeholder)
	assert.Nil(t, m2.transfer.err)
}

func TestInitTransferForm_DownloadNotAllowed(t *testing.T) {
	svc := &mockBinaryTransferService{}
	m := Model{
		currentState: "list",
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeFiles: svc,
		},
	}
	data := &model.BinaryData{ClientPath: ""}
	m2 := initTransferForm(m, modeDownload, data, "id1")
	assert.Equal(t, modeUpload, m2.transfer.mode)
	assert.Nil(t, m2.transfer.err)
	assert.Equal(t, "Назначение (куда сохранить)", m2.transfer.input.Placeholder)
}

func TestUpdateTransfer_CtrlU(t *testing.T) {
	m := Model{transfer: transferVM{mode: modeDownload, input: newInputField("")}}
	m2, _ := updateTransfer(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	assert.Equal(t, modeUpload, m2.transfer.mode)
	assert.Equal(t, "Источник (локальный файл)", m2.transfer.input.Placeholder)
}

func TestUpdateTransfer_CtrlD(t *testing.T) {
	data := &model.BinaryData{ClientPath: "/tmp/file"}
	m := Model{transfer: transferVM{mode: modeUpload, input: newInputField(""), data: data}}
	m2, _ := updateTransfer(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	assert.Equal(t, modeDownload, m2.transfer.mode)
	assert.Equal(t, "Назначение (куда сохранить)", m2.transfer.input.Placeholder)
}

func TestUpdateTransfer_CtrlD_CannotSwitch(t *testing.T) {
	data := &model.BinaryData{ClientPath: ""}
	m := Model{transfer: transferVM{mode: modeUpload, input: newInputField(""), data: data}}
	m2, _ := updateTransfer(m, tea.KeyMsg{Type: tea.KeyCtrlD})
	assert.Equal(t, modeUpload, m2.transfer.mode)
	require.Error(t, m2.transfer.err)
	assert.Contains(t, m2.transfer.err.Error(), "скачивание недоступно")
}

func TestUpdateTransfer_Enter(t *testing.T) {
	ti := newInputField("")
	ti.SetValue("")
	m := Model{transfer: transferVM{mode: modeUpload, input: ti}}
	m2, cmd := updateTransfer(m, tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
	require.Error(t, m2.transfer.err)
	assert.Equal(t, "укажите путь", m2.transfer.err.Error())
}

func TestUpdateTransfer_Esc(t *testing.T) {
	m := Model{currentState: "file_transfer", prevState: "menu", transfer: transferVM{input: newInputField(""), inFlight: false, data: &model.BinaryData{}}}
	m2, cmd := updateTransfer(m, tea.KeyMsg{Type: tea.KeyEsc})
	assert.Nil(t, cmd)
	assert.Equal(t, "menu", m2.currentState)
}

func TestUpdateTransfer_ProgressMsg(t *testing.T) {
	ch := make(chan int64)
	tv := transferVM{progCh: ch, total: 100, lastTickAt: time.Now().Add(-time.Second), inFlight: true}
	m := Model{transfer: tv}
	m2, cmd := updateTransfer(m, transferProgressMsg{sent: 150})
	require.NotNil(t, cmd)
	assert.Equal(t, int64(100), m2.transfer.sent)
}

func TestUpdateTransfer_DoneMsg(t *testing.T) {
	m := Model{transfer: transferVM{inFlight: true, err: errors.New("fail")}}
	m2, cmd := updateTransfer(m, transferDoneMsg{})
	assert.Nil(t, cmd)
	assert.False(t, m2.transfer.inFlight)
	assert.Nil(t, m2.transfer.err)
}

func TestUpdateTransfer_ErrorMsg(t *testing.T) {
	err := errors.New("boom")
	m := Model{transfer: transferVM{inFlight: true}}
	m2, cmd := updateTransfer(m, transferErrorMsg{err: err})
	assert.Nil(t, cmd)
	assert.False(t, m2.transfer.inFlight)
	assert.Equal(t, err, m2.transfer.err)
}

func TestUpdateTransfer_InputUpdate(t *testing.T) {
	ti := newInputField("")
	ti.Focus()
	m := Model{transfer: transferVM{input: ti}}
	m2, _ := updateTransfer(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.Equal(t, "a", m2.transfer.input.Value())
}

func TestRenderTransfer_NotInFlight(t *testing.T) {
	tv := transferVM{mode: modeUpload, input: newInputField(""), err: errors.New("boom")}
	m := Model{transfer: tv}
	out := renderTransfer(m)
	assert.Contains(t, out, "Загрузка файла на сервер")
	assert.Contains(t, out, "Enter: старт")
	assert.Contains(t, out, "Ошибка: boom")
}

func TestRenderTransfer_InFlightKnownTotal(t *testing.T) {
	tv := transferVM{mode: modeDownload, input: newInputField(""), inFlight: true, total: 100, sent: 50}
	m := Model{transfer: tv, termWidth: 80}
	out := renderTransfer(m)
	assert.Contains(t, out, "Скачивание файла с сервера")
	assert.Contains(t, out, "50 B / 100 B (50.0%)")
	assert.Contains(t, out, "Esc: отменить/назад")
}

func TestRenderTransfer_InFlightUnknownTotal(t *testing.T) {
	tv := transferVM{mode: modeUpload, input: newInputField(""), inFlight: true, total: 0, sent: 10}
	m := Model{transfer: tv, termWidth: 80}
	out := renderTransfer(m)
	assert.Contains(t, out, "Передано: 10 B (размер неизвестен)")
}
