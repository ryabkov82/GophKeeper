package contracts

import "context"

// DataType — перечисление типов объектов
type DataType int

const (
	TypeCredentials DataType = iota
	TypeNote
	TypeBinary
	TypeCard
)

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
