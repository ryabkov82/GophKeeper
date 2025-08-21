package forms

import (
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// Adapt преобразует доменную сущность в форму.
// Если сущность уже реализует FormEntity, она возвращается как есть.
func Adapt(v interface{}) (FormEntity, error) {
	if fe, ok := v.(FormEntity); ok {
		return fe, nil
	}
	switch e := v.(type) {
	case *model.BankCard:
		return &BankCardAdapter{BankCard: e}, nil
	case *model.Credential:
		return &CredentialAdapter{Credential: e}, nil
	case *model.TextData:
		return &TextDataAdapter{TextData: e}, nil
	case *model.BinaryData:
		return &BinaryDataAdapter{BinaryData: e}, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}
