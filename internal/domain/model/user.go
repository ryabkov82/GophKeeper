package model

// User представляет сущность пользователя в системе.
//
// Поля:
//   - ID: уникальный идентификатор пользователя (например, UUID);
//   - Login: имя пользователя для аутентификации;
//   - PasswordHash: хеш пароля пользователя;
//   - Salt: соль, использованная при хешировании пароля.
type User struct {
	ID           string
	Login        string
	PasswordHash string
	Salt         string
}
