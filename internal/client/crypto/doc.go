// Package crypto предоставляет клиентские криптографические утилиты:
//
// - генерацию симметричных ключей из пароля и соли с помощью Argon2id;
// - шифрование и расшифровку данных с использованием AES-GCM.
//
// Основное предназначение — формирование и использование ключа для шифрования приватных данных
// перед отправкой их на сервер и после получения с сервера.
//
// Пример использования:
//
//	salt := []byte("случайная_соль_с_сервера")
//	key, err := crypto.DeriveKey("пароль_пользователя", salt)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ciphertext, err := crypto.EncryptAESGCM([]byte("секретные данные"), key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	plaintext, err := crypto.DecryptAESGCM(ciphertext, key)
//	if err != nil {
//	    log.Fatal(err)
//	}
package crypto
