package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/adapters"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
)

// Определяем структуру, которая содержит все нужные сервисы для модели
type ModelServices struct {
	Auth       contracts.AuthService
	Credential contracts.CredentialService
	// Добавляй сюда другие интерфейсы по необходимости
}

// Model - основная модель приложения, реализующая tea.Model
type Model struct {
	// Состояния интерфейса
	currentState string // "menu", "login", "register", "list", "view", "edit"

	// Главное меню
	menuItems  []menuItem
	menuCursor int

	// Формы ввода
	inputs       []textinput.Model
	focusedInput int

	// Данные приложения
	authService contracts.AuthService
	ctx         context.Context
	registerErr error // Добавляем поле для ошибок
	loginErr    error

	currentType contracts.DataType   // какой тип данных сейчас выбран
	listItems   []contracts.ListItem // универсальный список
	listCursor  int

	// map: DataType -> DataService
	services map[contracts.DataType]contracts.DataService

	// для редактирования — generic формы
	editEntity interface{} // храним сущность (например *model.Credential)

	lastError error // Добавляем поле для ошибок
}

// Добавляем сообщения для системы
type RegisterSuccessMsg struct{}
type RegisterFailedMsg struct{ Err error }
type listLoadedMsg struct{ items []contracts.ListItem }
type errMsg struct{ err error }

// menuItem - элемент меню
type menuItem struct {
	title       string
	description string
}

// NewModel создает новую модель приложения
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
		},
	}
}

// Init начальная команда при запуске
func (m Model) Init() tea.Cmd {
	return nil
}

// Update обработка сообщений и обновление состояния
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.lastError = msg.err
			return m, nil

		default:
			return updateViewData(m, msg)
		}
	case "edit_new": // <- здесь обрабатываем редактирование новой записи
		return updateEdit(m, msg)
	default:
		return m, nil
	}
}

// View рендеринг интерфейса
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
	default:
		return ""
	}
}

func (m *Model) loadList() tea.Cmd {
	return func() tea.Msg {
		items, err := m.services[m.currentType].List(m.ctx)
		if err != nil {
			return errMsg{err}
		}
		return listLoadedMsg{items}
	}
}

// Стили интерфейса
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	//cursorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	hintStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	activeFieldStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	inactiveFieldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)
