package auth

type AuthManager struct {
	token      string       // Текущий токен (в памяти)
	tokenStore TokenStorage // Постоянное хранилище (файл, keychain и т.д.)
}

type TokenStorage interface {
	Save(token string) error
	Load() (string, error)
	Clear() error
}

func NewAuthManager(store TokenStorage) *AuthManager {
	return &AuthManager{
		tokenStore: store,
	}
}

func (a *AuthManager) SetToken(token string) error {
	a.token = token
	return a.tokenStore.Save(token)
}

func (a *AuthManager) GetToken() string {
	if a.token == "" {
		token, _ := a.tokenStore.Load() // Игнорируем ошибку, если токена нет
		a.token = token
	}
	return a.token
}

func (a *AuthManager) Clear() error {
	a.token = ""
	return a.tokenStore.Clear()
}
