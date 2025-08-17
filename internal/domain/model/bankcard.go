package model

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

// Глобальная переменная для подмены
var timeNow = time.Now

// BankCard — модель для хранения данных банковской карты.
// Все чувствительные поля (номер карты, срок действия, CVV, имя владельца)
// должны храниться в зашифрованном виде (например, base64).
type BankCard struct {
	ID             string    `db:"id"`              // Уникальный идентификатор карты (UUID)
	UserID         string    `db:"user_id"`         // Идентификатор пользователя-владельца карты
	Title          string    `db:"title"`           // Название или ярлык карты (например, "Рабочая карта")
	CardholderName string    `db:"cardholder_name"` // Имя держателя карты, как указано на карте
	CardNumber     string    `db:"card_number"`     // Номер карты (обычно 16 цифр)
	ExpiryDate     string    `db:"expiry_date"`     // Срок действия карты в формате MM/YY
	CVV            string    `db:"cvv"`             // Код безопасности карты (3 или 4 цифры)
	Metadata       string    `db:"metadata"`        // Дополнительные данные в формате JSON или свободный текст
	CreatedAt      time.Time `db:"created_at"`      // Время создания записи
	UpdatedAt      time.Time `db:"updated_at"`      // Время последнего обновления записи
}

// ValidateCardNumber проверяет номер карты
func (b *BankCard) ValidateCardNumber() error {
	return validateCardNumber(b.CardNumber)
}

// validateCardNumber проверяет номер карты по следующим критериям:
// 1. Длина от 12 до 19 цифр (по стандарту ISO/IEC 7812)
// 2. Соответствие алгоритму Луна
// 3. Допустимые префиксы основных платежных систем
func validateCardNumber(number string) error {
	cleaned := strings.ReplaceAll(number, " ", "")
	if len(cleaned) < 12 || len(cleaned) > 19 {
		return errors.New("card number must contain 12 to 19 digits")
	}

	if !luhnCheck(cleaned) {
		return errors.New("invalid card number (Luhn check failed)")
	}

	if !checkPaymentSystem(cleaned) {
		return errors.New("unsupported card type")
	}

	return nil
}

// luhnCheck реализует алгоритм Луна для проверки номера карты
func luhnCheck(number string) bool {
	sum := 0
	parity := len(number) % 2

	for i, char := range number {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}

// checkPaymentSystem проверяет префиксы основных платежных систем
func checkPaymentSystem(number string) bool {
	// Visa: начинается на 4
	if matched, _ := regexp.MatchString(`^4[0-9]{12}(?:[0-9]{3})?$`, number); matched {
		return true
	}

	// Mastercard: начинается на 51-55 или 2221-2720
	if matched, _ := regexp.MatchString(`^(5[1-5][0-9]{14}|2(22[1-9][0-9]{12}|2[3-9][0-9]{13}|[3-6][0-9]{14}|7[0-1][0-9]{13}|720[0-9]{12}))$`, number); matched {
		return true
	}

	// Мир: начинается на 2200-2204
	if matched, _ := regexp.MatchString(`^220[0-4][0-9]{12}$`, number); matched {
		return true
	}

	// American Express: начинается на 34 или 37
	if matched, _ := regexp.MatchString(`^3[47][0-9]{13}$`, number); matched {
		return true
	}

	// Можно добавить другие платежные системы по необходимости

	return false
}

func validateExpiryDate(expiry string) error {
	now := timeNow()
	// Удаляем все пробелы и проверяем длину
	expiry = strings.ReplaceAll(expiry, " ", "")
	if len(expiry) != 5 || expiry[2] != '/' {
		return errors.New("invalid expiry date format, use MM/YY")
	}

	parts := strings.Split(expiry, "/")
	if len(parts) != 2 {
		return errors.New("invalid expiry date format")
	}

	month, err1 := strconv.Atoi(parts[0])
	year, err2 := strconv.Atoi(parts[1])

	// Проверка ошибок парсинга
	if err1 != nil || err2 != nil {
		return errors.New("month and year must be numbers")
	}

	// Проверка месяца
	if month < 1 || month > 12 {
		return errors.New("invalid month, must be between 01 and 12")
	}

	// Проверка года (допустим диапазон текущий год ±5 лет)
	currentYear := now.Year() % 100
	currentMonth := now.Month()

	if year < currentYear || year > currentYear+5 {
		return errors.New("invalid year, card expired or too far in future")
	}

	// Проверка на истекший срок
	if year == currentYear && month < int(currentMonth) {
		return errors.New("card has expired")
	}

	return nil
}

// Реализация интерфейса tui.FormEntity

// FormFields возвращает описание полей формы для редактирования BankCard
func (b *BankCard) FormFields() []forms.FormField {
	return []forms.FormField{
		{
			Label:       "Title",
			Value:       b.Title,
			MaxLength:   150,
			InputType:   "text",
			Placeholder: "Название карты (например, Основная карта)",
		},
		{
			Label:       "Cardholder Name",
			Value:       b.CardholderName,
			MaxLength:   150,
			InputType:   "text",
			Placeholder: "Имя держателя карты",
		},
		{
			Label:       "Card Number",
			Value:       b.CardNumber,
			InputType:   "text",
			Placeholder: "Номер карты (XXXX XXXX XXXX XXXX)",
			Mask:        "#### #### #### ####",
		},
		{
			Label:       "Expiry Date",
			Value:       b.ExpiryDate,
			InputType:   "text",
			Placeholder: "MM/YY",
			Mask:        "##/##",
		},
		{
			Label:       "CVV",
			Value:       b.CVV,
			InputType:   "password",
			Placeholder: "XXX",
			Mask:        "###",
		},
		{
			Label:       "Metadata",
			Value:       b.Metadata,
			InputType:   "multiline",
			Fullscreen:  false,
			Placeholder: "Дополнительная информация о карте",
		},
	}
}

// UpdateFromFields обновляет BankCard по значениям из формы
func (b *BankCard) UpdateFromFields(fields []forms.FormField) error {
	if len(fields) != 6 {
		return errors.New("unexpected number of fields")
	}

	// Базовая валидация
	if strings.TrimSpace(fields[0].Value) == "" {
		return errors.New("title cannot be empty")
	}

	cardNumber := strings.ReplaceAll(fields[2].Value, " ", "")
	if err := validateCardNumber(cardNumber); err != nil {
		return err
	}

	expiryDate := fields[3].Value
	if err := validateExpiryDate(expiryDate); err != nil {
		return err
	}

	cvv := fields[4].Value
	if _, err := strconv.Atoi(cvv); err != nil || len(cvv) != 3 {
		return errors.New("CVV must be 3 digits")
	}

	// Обновляем поля
	b.Title = fields[0].Value
	b.CardholderName = fields[1].Value
	b.CardNumber = cardNumber
	b.ExpiryDate = expiryDate
	b.CVV = fields[4].Value
	b.Metadata = fields[5].Value
	b.UpdatedAt = time.Now()

	return nil
}

func (b *BankCard) GetID() string   { return b.ID }
func (b *BankCard) SetID(id string) { b.ID = id }
