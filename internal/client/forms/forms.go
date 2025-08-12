package forms

type FormField struct {
	Label       string
	Value       string
	InputType   string // "text", "password" и т.д.
	Placeholder string
}

type FormEntity interface {
	FormFields() []FormField
	UpdateFromFields(fields []FormField) error
}
