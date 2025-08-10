// Package jwtutils предоставляет функционал для создания и проверки JWT токенов,
// используемых для аутентификации пользователей.
//
// В пакете определён тип TokenManager, который инкапсулирует секретный ключ и время жизни токена (TTL).
// С его помощью можно создавать JWT с пользовательскими claims (идентификатором пользователя и логином),
// а также разбирать и валидировать входящие JWT.
//
// Основные возможности пакета:
//   - Генерация JWT с пользовательскими claims и сроком действия.
//   - Проверка подписи и валидности JWT.
//   - Парсинг claims из JWT.
//
// Пакет предназначен для использования в сервисах, где необходима безопасная аутентификация
// и передача информации о пользователе в виде JWT.
//
// Пример создания менеджера и генерации токена:
//
//	tm := jwtutils.New("секретный_ключ", time.Hour*24)
//	token, err := tm.GenerateToken(userID, login)
//
// Пример парсинга и проверки токена:
//
//	claims, err := tm.ParseToken(tokenString)
//	if err != nil {
//	    // обработка ошибки (например, невалидный токен)
//	}
//	userID := claims["sub"].(string)
//
// Ошибки:
//   - ErrTokenInvalid возвращается, если токен невалиден (просрочен, поврежден или имеет неверный формат).
//   - jwt.ErrSignatureInvalid возвращается при некорректной подписи токена.
package jwtutils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

var ErrTokenInvalid = errors.New("token is invalid")

// New создает новый менеджер токенов
func New(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// GenerateToken создает JWT для указанного пользователя
func (tm *TokenManager) GenerateToken(userID, login string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"login": login,
		"exp":   time.Now().Add(tm.ttl).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secret)
}

// ParseToken разбирает и проверяет JWT токен из строки tokenStr.
// Возвращает claims (набор данных токена) в виде jwt.MapClaims, если токен валиден.
// Проверяет алгоритм подписи, срок действия и корректность токена.
//
// Параметры:
//   - tokenStr: JWT токен в виде строки.
//
// Возвращает:
//   - jwt.MapClaims: распарсенные claims токена, если он валиден.
//   - error: ошибку, если токен невалиден, подпись некорректна или произошла другая ошибка парсинга.
//
// Возможные ошибки:
//   - jwt.ErrSignatureInvalid: если алгоритм подписи не соответствует ожиданиям или подпись некорректна.
//   - ErrTokenInvalid: если токен невалиден (например, просрочен или поврежден).
//   - Другие ошибки, возникающие при парсинге токена.
//
// Использование:
//
//	claims, err := tm.ParseToken(tokenString)
//	if err != nil {
//	    // обработка ошибки
//	}
//	userID := claims["sub"].(string)
func (tm *TokenManager) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return tm.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// Secret возвращает секретный ключ, используемый для подписи токенов.
func (tm *TokenManager) Secret() []byte {
	return tm.secret
}
