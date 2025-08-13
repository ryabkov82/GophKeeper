package crypto

import (
	"errors"

	"golang.org/x/crypto/argon2"
)

// Argon2Params содержит параметры для функции Argon2id,
// используемой для генерации ключа из пароля и соли.
type Argon2Params struct {
	Time    uint32 // Количество проходов алгоритма
	Memory  uint32 // Используемая память в KiB
	Threads uint8  // Количество параллельных потоков
	KeyLen  uint32 // Длина выходного ключа в байтах
}

// DefaultParams — рекомендуемые параметры Argon2id,
// обеспечивающие баланс безопасности и производительности.
var DefaultParams = Argon2Params{
	Time:    1,
	Memory:  64 * 1024, // 64 MiB
	Threads: 4,
	KeyLen:  32, // 256 бит
}

// DeriveKey генерирует ключ из пароля и соли с помощью Argon2id,
// возвращая сам ключ и параметры, которые были использованы.
func DeriveKey(password string, salt []byte) ([]byte, Argon2Params, error) {
	if len(salt) == 0 {
		return nil, Argon2Params{}, errors.New("salt cannot be empty")
	}

	params := DefaultParams

	// Если salt короче, чем нужно, можно либо ошибку, либо дополнить.
	// Но здесь считаем, что salt приходит с сервера корректный.

	key := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)
	return key, params, nil
}
