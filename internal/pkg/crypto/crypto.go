// Package crypto предоставляет утилиты для безопасной обработки паролей,
// включая их хеширование с солью и проверку.
//
// В основе реализации используется алгоритм SHA-256 и случайно
// генерируемая соль длиной 16 байт, закодированная в base64.
// Такой подход усложняет атаку перебора (brute-force) и делает невозможным
// использование заранее подготовленных таблиц (rainbow tables).
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// HashPassword принимает пароль в виде строки и возвращает его хеш и соль.
//
// Алгоритм:
//  1. Проверяет, что пароль не пустой.
//  2. Генерирует криптографически стойкую соль длиной 16 байт с помощью crypto/rand.
//  3. Кодирует соль в base64.
//  4. Вычисляет SHA-256 хеш строки password+salt.
//  5. Кодирует хеш в base64 и возвращает его вместе с солью.
//
// Параметры:
//   - password: исходный пароль пользователя.
//
// Возвращает:
//   - hash: строка с base64-представлением SHA-256 хеша.
//   - salt: строка с base64-представлением соли.
//   - err: ошибка генерации случайных байт или пустой пароль.
//
// Пример:
//
//	hash, salt, err := crypto.HashPassword("mysecurepassword")
//	if err != nil {
//	    log.Fatal(err)
//	}
func HashPassword(password string) (hash, salt string, err error) {
	if password == "" {
		return "", "", errors.New("password is empty")
	}

	// Генерируем соль — 16 байт случайных данных
	saltBytes := make([]byte, 16)
	_, err = rand.Read(saltBytes)
	if err != nil {
		return "", "", err
	}
	salt = base64.StdEncoding.EncodeToString(saltBytes)

	// Хешируем пароль + соль
	hashBytes := sha256.Sum256([]byte(password + salt))
	hash = base64.StdEncoding.EncodeToString(hashBytes[:])

	return hash, salt, nil
}

// VerifyPassword проверяет соответствие пароля переданным хешу и соли.
//
// Алгоритм:
//  1. Конкатенирует пароль с солью.
//  2. Вычисляет SHA-256 хеш результата.
//  3. Кодирует полученный хеш в base64.
//  4. Сравнивает его с переданным значением hash.
//
// Параметры:
//   - password: введённый пользователем пароль.
//   - hash: сохранённый хеш пароля (base64).
//   - salt: сохранённая соль (base64).
//
// Возвращает:
//   - true, если пароль корректен.
//   - false, если пароль не совпадает.
//
// Пример:
//
//	if crypto.VerifyPassword(inputPassword, storedHash, storedSalt) {
//	    fmt.Println("Пароль верный")
//	} else {
//	    fmt.Println("Пароль неверный")
//	}
func VerifyPassword(password, hash, salt string) bool {
	hashBytes := sha256.Sum256([]byte(password + salt))
	computedHash := base64.StdEncoding.EncodeToString(hashBytes[:])
	return computedHash == hash
}
