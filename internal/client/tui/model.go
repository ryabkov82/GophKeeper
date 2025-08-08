package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model - основная модель приложения, реализующая tea.Model
type Model struct {
	// Состояния интерфейса
	currentState string // "menu", "login", "register", "view_data"

	// Главное меню
	menuItems  []menuItem
	menuCursor int

	// Формы ввода
	inputs       []textinput.Model
	focusedInput int

	// Данные приложения
	services    *AppServices
	ctx         context.Context
	registerErr error // Добавляем поле для ошибок
	loginErr    error
}

// Добавляем сообщения для системы
type RegisterSuccessMsg struct{}
type RegisterFailedMsg struct{ Err error }

// menuItem - элемент меню
type menuItem struct {
	title       string
	description string
}

// User - данные пользователя
type User struct {
	Username string
	Token    string
}

// Credential - хранимые учетные данные
type Credential struct {
	Type     string
	Username string
	Password string
	Metadata string
}

// NewModel создает новую модель приложения
func NewModel(ctx context.Context, services *AppServices) *Model {

	return &Model{
		currentState: "menu",
		menuItems: []menuItem{
			{"Login", "Войти в систему"},
			{"Register", "Зарегистрироваться"},
			{"View Data", "Просмотреть сохраненные данные"},
			{"Exit", "Выйти из приложения"},
		},
		inputs:       make([]textinput.Model, 0),
		focusedInput: 0,
		ctx:          ctx,
		services:     services,
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
	case "view_data":
		return updateViewData(m, msg)
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
	case "view_data":
		return renderViewData(m)
	default:
		return ""
	}
}

// Стили интерфейса
var (
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	errorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	selectedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	normalStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	hintStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	activeFieldStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	inactiveFieldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)
