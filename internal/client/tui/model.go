package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ryabkov82/gophkeeper/internal/client/forms"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/adapters"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
)

// Определяем структуру, которая содержит все нужные сервисы для модели
type ModelServices struct {
	Auth       contracts.AuthService       // Сервис аутентификации
	Credential contracts.CredentialService // Сервис управления учетными данными
	Bankcard   contracts.BankCardService   // Сервис управления банковскими картами
	// Добавляй сюда другие интерфейсы по необходимости
}

// formWidget представляет отдельное поле формы, может быть обычным input или textarea.
type formWidget struct {
	isTextarea  bool
	input       textinput.Model
	textarea    textarea.Model
	field       forms.FormField
	maskedInput MaskedInput // ← новое поле для работы с маской
}

type MaskedInput struct {
	Mask      string
	Raw       []rune
	CursorPos int
}

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

func isPlaceholder(r rune) bool {
	return r == '#'
}

// Проверка вводимого символа по маске
func isRuneAllowedForMaskChar(maskChar, input rune) bool {
	switch maskChar {
	case '#':
		return input >= '0' && input <= '9'
	default:
		return false
	}
}

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

func (m *MaskedInput) InsertString(s string) {
	for _, r := range s {
		m.InsertRune(r)
	}
}

func (m *MaskedInput) Backspace() {
	for i := m.CursorPos - 1; i >= 0; i-- {
		if isPlaceholder(rune(m.Mask[i])) && m.Raw[i] != ' ' {
			m.Raw[i] = ' '
			m.CursorPos = i
			break
		}
	}
}

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

func (m *MaskedInput) Home() {
	for i, r := range m.Mask {
		if isPlaceholder(r) {
			m.CursorPos = i
			break
		}
	}
}

func (m *MaskedInput) End() {
	last := 0
	for i, r := range m.Mask {
		if isPlaceholder(r) && m.Raw[i] != ' ' {
			last = i
		}
	}
	m.CursorPos = last
}

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

func (m *MaskedInput) MoveLeft() {
	for i := m.CursorPos - 1; i >= 0; i-- {
		if isPlaceholder(rune(m.Mask[i])) {
			m.CursorPos = i
			break
		}
	}
}

func (m *MaskedInput) MoveRight() {
	for i := m.CursorPos + 1; i < len(m.Mask); i++ {
		if isPlaceholder(rune(m.Mask[i])) {
			m.CursorPos = i
			break
		}
	}
}

// setFocus устанавливает фокус на виджет или снимает его.
func (w *formWidget) setFocus(focused bool) {
	if w.isTextarea {
		if focused {
			w.textarea.Focus()
		} else {
			w.textarea.Blur()
		}
	} else {
		if focused {
			w.input.Focus()
		} else {
			w.input.Blur()
		}
	}
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
		},
	}
}

// Init начальная команда при запуске
func (m Model) Init() tea.Cmd {
	return nil
}

// Update обрабатывает входящие сообщения и обновляет состояние модели.
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
			m.listErr = msg.err
			return m, nil

		default:
			return updateViewData(m, msg)
		}
	case "edit", "edit_new":
		return updateEdit(m, msg)
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

// Стили интерфейса (для заголовков, ошибок, активных/неактивных полей и подсказок).
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
