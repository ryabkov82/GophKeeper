package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/adapters"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
)

// Определяем структуру, которая содержит все нужные сервисы для модели
type ModelServices struct {
	Auth       contracts.AuthService       // Сервис аутентификации
	Credential contracts.CredentialService // Сервис управления учетными данными
	Bankcard   contracts.BankCardService   // Сервис управления банковскими картами
	TextData   contracts.TextDataService   // Сервис управления текстовыми данными
	BinaryData contracts.BinaryDataService // Сервис управления бинарными данными
	// Добавляй сюда другие интерфейсы по необходимости
}

// Model - основная модель приложения, реализующая tea.Model
type Model struct {
	// Состояния интерфейса
	currentState string // "menu", "login", "register", "list", "view", "edit"

	// Главное меню
	menuItems  []menuItem // элементы главного меню
	menuCursor int        // текущая позиция в меню

	// Формы ввода
	inputs       []textinput.Model // срез обычных полей ввода
	focusedInput int               // индекс текущего фокусного поля

	// Данные приложения
	authService contracts.AuthService // сервис аутентификации
	ctx         context.Context       // контекст приложения
	registerErr error                 // ошибка регистрации
	loginErr    error                 // ошибка логина

	currentType contracts.DataType   // какой тип данных сейчас выбран
	listItems   []contracts.ListItem // универсальный список элементов
	listCursor  int                  // индекс выбранного элемента списка
	listErr     error                // ошибка загрузки списка

	// map: DataType -> DataService
	services map[contracts.DataType]contracts.DataService // карта сервисов для каждого типа данных

	// для редактирования — generic формы
	editEntity interface{}  // храним сущность (например *model.Credential)
	widgets    []formWidget // виджеты формы редактирования
	editErr    error        // ошибка редактирования

	fullscreenWidget *formWidget // ссылка на виджет в режиме fullscreen
	prevState        string      // сохраняем состояние перед fullscreen
	fullscreenErr    error       // ошибка редактирования элемента в режиме fullscreen

	termWidth  int // ширина терминала
	termHeight int // высота терминала

	transfer transferVM //структура для реализации передачи файлов
}

// Добавляем сообщения для системы
type RegisterSuccessMsg struct{}
type RegisterFailedMsg struct{ Err error }
type listLoadedMsg struct{ items []contracts.ListItem }
type errMsg struct{ err error }

// menuItem описывает элемент главного меню.
type menuItem struct {
	title       string // заголовок пункта
	description string // описание/подсказка
}

// NewModel создаёт новую модель приложения с заданными сервисами и контекстом.
func NewModel(ctx context.Context, svcs ModelServices) *Model {

	return &Model{
		currentState: "menu",
		menuItems: []menuItem{
			{"Login", "Войти в систему"},
			{"Register", "Зарегистрироваться"},
			{"Credentials", "Учётные данные"},
			{"Notes", "Текстовые заметки"},
			{"Files", "Бинарные файлы"},
			{"Cards", "Банковские карты"},
			{"Exit", "Выйти из приложения"},
		},
		inputs:       make([]textinput.Model, 0),
		focusedInput: 0,
		ctx:          ctx,
		authService:  svcs.Auth,
		services: map[contracts.DataType]contracts.DataService{
			contracts.TypeCredentials: adapters.NewCredentialAdapter(svcs.Credential),
			contracts.TypeCards:       adapters.NewBankCardAdapter(svcs.Bankcard),
			contracts.TypeNotes:       adapters.NewTextDataAdapter(svcs.TextData),
			contracts.TypeFiles:       adapters.NewBinaryDataAdapter(svcs.BinaryData),
		},
	}
}

// Init начальная команда при запуске
func (m Model) Init() tea.Cmd {
	return nil
}

// Update обрабатывает входящие сообщения и обновляет состояние модели.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// --- Обработка ресайза терминала ---
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height

		if m.fullscreenWidget != nil {
			m.fullscreenWidget.textarea.SetWidth(m.termWidth - 2)
			m.fullscreenWidget.textarea.SetHeight(m.termHeight - 8)
		}
		if m.widgets != nil {
			for i, w := range m.widgets {
				if w.isTextarea {
					m.widgets[i].textarea.SetWidth(m.termWidth - 2)
				}
			}
		}
		return m, nil
	}

	switch m.currentState {
	case "menu":
		return updateMenu(m, msg)
	case "login":
		return updateLogin(m, msg)
	case "loginSuccess":
		return updateLoginSuccess(m, msg)
	case "register":
		return updateRegister(m, msg)
	case "registerSuccess":
		return updateRegisterSuccess(m, msg)
	case "list":
		// Обрабатываем сообщения listLoadedMsg и errMsg
		switch msg := msg.(type) {
		case listLoadedMsg:
			m.listItems = msg.items
			m.listCursor = 0
			return m, nil

		case errMsg:
			// Сохраняем ошибку в модель, чтобы показать пользователю
			m.listErr = msg.err
			return m, nil

		default:
			return updateViewData(m, msg)
		}
	case "edit", "edit_new":
		return updateEdit(m, msg)
	case "fullscreen_edit":
		return updateFullscreenForm(m, msg)
	case "file_transfer":
		return updateTransfer(m, msg)
	default:
		return m, nil
	}
}

// View рендерит текущее состояние интерфейса в строку для вывода в терминал.
func (m Model) View() string {
	switch m.currentState {
	case "menu":
		return renderMenu(m)
	case "login":
		return renderLogin(m)
	case "loginSuccess":
		return renderLoginSuccess(m)
	case "register":
		return renderRegister(m)
	case "registerSuccess":
		return renderRegisterSuccess(m)
	case "list":
		return renderList(m)
	case "edit", "edit_new":
		return renderEditForm(m)
	case "fullscreen_edit":
		return renderFullscreenForm(m)
	case "file_transfer":
		return renderTransfer(m)
	default:
		return ""
	}
}

// loadList возвращает команду для загрузки списка элементов текущего типа.
func (m *Model) loadList() tea.Cmd {
	return func() tea.Msg {
		items, err := m.services[m.currentType].List(m.ctx)
		if err != nil {
			return errMsg{err}
		}
		return listLoadedMsg{items}
	}
}

// ExtractFields извлекает значения из всех виджетов формы (inputs и textarea) в срез FormField.
func (m Model) ExtractFields() []forms.FormField {
	result := make([]forms.FormField, 0, len(m.widgets))

	for _, w := range m.widgets {
		f := w.field
		if w.isTextarea {
			f.Value = w.textarea.Value()
		} else {
			f.Value = w.input.Value()
		}
		result = append(result, f)
	}

	return result
}
