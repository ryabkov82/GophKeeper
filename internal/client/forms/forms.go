// Package forms содержит структуры и интерфейсы для описания форм ввода
// и связи пользовательских форм с доменными сущностями.
//
// Он используется в TUI для генерации форм редактирования/создания сущностей,
// обработки ввода и валидации данных.
package forms

// FormField описывает одно поле формы для редактирования сущности.
// Label — текст метки поля,
// Value — текущее значение,
// InputType — тип ввода ("text", "password", "multiline" и т.д.),
// Placeholder — текст-подсказка, отображаемый в пустом поле.
type FormField struct {
	Label       string
	Value       string
	InputType   string // "text", "password" и т.д.
	Placeholder string
	Mask        string
	MaxLength   int  // для ограничения длины ввода
	Fullscreen  bool // если true — поле может редактироваться в полноэкранном режиме
	ReadOnly    bool // если true — поле доступно только для чтения
}

// FormEntity интерфейс, который должна реализовывать любая сущность,
// для которой генерируется форма редактирования.
// FormFields возвращает описание всех полей формы.
// UpdateFromFields обновляет сущность по значениям из формы.
type FormEntity interface {
	FormFields() []FormField
	UpdateFromFields(fields []FormField) error
}

// Identifiable интерфейс для сущностей, у которых есть идентификатор.
// Позволяет получать и устанавливать ID сущности.
type Identifiable interface {
	GetID() string
	SetID(id string)
}
