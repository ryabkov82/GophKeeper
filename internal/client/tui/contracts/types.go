package contracts

import "context"

// DataType — перечисление типов объектов
type DataType int

const (
	TypeCredentials DataType = iota
	TypeNotes
	TypeFiles
	TypeCards
)

// String возвращает читаемое название типа данных DataType.
// Используется для отображения в интерфейсе и логах.
//
// Примеры возврата:
//   TypeCredentials -> "Credentials"
//   TypeNotes       -> "Notes"
//   TypeFiles       -> "Files"
//   TypeCards       -> "Cards"
//   любое другое значение -> "Unknown"
func (dt DataType) String() string {
	switch dt {
	case TypeCredentials:
		return "Credentials"
	case TypeNotes:
		return "Notes"
	case TypeFiles:
		return "Files"
	case TypeCards:
		return "Cards"
	default:
		return "Unknown"
	}
}

// ListItem — единица для универсального списка
type ListItem struct {
	ID       string
	Title    string
	Subtitle string // короткая подсказка (например, login или дата)
	Type     DataType
}

// DataService — общий интерфейс для CRUD-операций для любого типа
type DataService interface {
	List(ctx context.Context) ([]ListItem, error)            // вернуть элементы списка
	Get(ctx context.Context, id string) (interface{}, error) // вернуть полную сущность (тип-specific)
	Create(ctx context.Context, v interface{}) error
	Update(ctx context.Context, id string, v interface{}) error
	Delete(ctx context.Context, id string) error
}
