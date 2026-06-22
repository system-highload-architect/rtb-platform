package fixedpoint

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Money представляет денежную сумму в минимальных неделимых единицах (копейки, центы и т.д.).
// Нулевое значение — 0. Допустимы отрицательные значения.
type Money int64

// Константы для масштабирования (степень 10).
const (
	scale  = 100            // количество копеек в единице валюты
	scaleF = float64(scale) // для преобразований с float
)

// Стандартные ошибки.
var (
	ErrOverflow      = errors.New("fixedpoint: overflow")
	ErrInsufficient  = errors.New("fixedpoint: insufficient funds")
	ErrInvalidFormat = errors.New("fixedpoint: invalid format")
)

// ──────────── Конструкторы ────────────

// NewFromInt64 создаёт Money из целых единиц валюты (рублей/долларов).
func NewFromInt64(units int64) Money {
	return Money(units * scale)
}

// NewFromFloat64 создаёт Money из float64 (рублей/долларов), округляя до копейки.
// Отрицательные значения допустимы.
func NewFromFloat64(amount float64) (Money, error) {
	cents := amount * scaleF
	// Округляем до ближайшего целого
	rounded := roundFloat(cents)
	if rounded > 1<<63-1 || rounded < -1<<63 {
		return 0, ErrOverflow
	}
	return Money(rounded), nil
}

// ParseMoney разбирает строку вида "123.45" или "-67.89" в Money.
func ParseMoney(s string) (Money, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidFormat
	}
	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	parts := strings.SplitN(s, ".", 2)
	units, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid units %q", ErrInvalidFormat, parts[0])
	}

	var cents int64
	if len(parts) == 2 {
		// Дробная часть: дополняем до двух цифр
		centsStr := parts[1]
		if len(centsStr) == 1 {
			centsStr += "0"
		} else if len(centsStr) > 2 {
			centsStr = centsStr[:2] // обрезаем до копеек
		}
		cents, err = strconv.ParseInt(centsStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: invalid cents %q", ErrInvalidFormat, parts[1])
		}
	}

	amount := units*scale + cents
	if negative {
		amount = -amount
	}
	return Money(amount), nil
}

// MustParseMoney как ParseMoney, но паникует при ошибке.
func MustParseMoney(s string) Money {
	m, err := ParseMoney(s)
	if err != nil {
		panic(err)
	}
	return m
}

// ──────────── Арифметические операции ────────────

// Add возвращает m + other. При переполнении возвращает ошибку.
func (m Money) Add(other Money) (Money, error) {
	result := int64(m) + int64(other)
	if (result > int64(m)) != (int64(other) > 0) {
		return 0, ErrOverflow
	}
	return Money(result), nil
}

// Sub возвращает m - other. При недостатке средств (если m < other) возвращает ErrInsufficient.
// При переполнении возвращает ErrOverflow.
func (m Money) Sub(other Money) (Money, error) {
	if m < other {
		return 0, ErrInsufficient
	}
	result := int64(m) - int64(other)
	// Проверка переполнения не нужна, т.к. m >= other и знаки известны,
	// но для полной безопасности:
	if (result < int64(m)) != (int64(other) > 0) {
		// теоретически невозможно из-за проверки выше, но оставим
		return 0, ErrOverflow
	}
	return Money(result), nil
}

// Mul умножает деньги на целое число (например, на количество показов).
// Переполнение контролируется.
func (m Money) Mul(factor int64) (Money, error) {
	if factor == 0 || m == 0 {
		return 0, nil
	}
	result := int64(m) * factor
	// Проверка переполнения: деление результата обратно должно дать m (для factor != 0)
	if result/factor != int64(m) {
		return 0, ErrOverflow
	}
	return Money(result), nil
}

// Div делит деньги на целое число с округлением вниз.
// Возвращает ошибку, если divisor == 0.
func (m Money) Div(divisor int64) (Money, error) {
	if divisor == 0 {
		return 0, errors.New("fixedpoint: division by zero")
	}
	// Округление вниз (отрицательные числа тоже округляются вниз)
	q := int64(m) / divisor
	r := int64(m) % divisor
	// Если есть остаток и результат отрицательный, уменьшаем на 1
	if r != 0 && (int64(m) < 0) != (divisor < 0) {
		q--
	}
	return Money(q), nil
}

// MulFloat умножает деньги на float64 коэффициент и округляет до копеек.
// Используется для применения скоринговых коэффициентов.
func (m Money) MulFloat(coeff float64) (Money, error) {
	if coeff == 0 || m == 0 {
		return 0, nil
	}
	result := float64(m) * coeff
	rounded := roundFloat(result)
	if rounded > 1<<63-1 || rounded < -1<<63 {
		return 0, ErrOverflow
	}
	return Money(rounded), nil
}

// ──────────── Сравнение и вспомогательные методы ────────────

// Cmp возвращает -1 если m < other, 0 если равны, 1 если m > other.
func (m Money) Cmp(other Money) int {
	switch {
	case m < other:
		return -1
	case m > other:
		return 1
	default:
		return 0
	}
}

// IsZero возвращает true, если сумма равна нулю.
func (m Money) IsZero() bool {
	return m == 0
}

// Abs возвращает абсолютное значение.
func (m Money) Abs() Money {
	if m < 0 {
		return -m
	}
	return m
}

// Sign возвращает 1 для положительных, -1 для отрицательных, 0 для нуля.
func (m Money) Sign() int {
	switch {
	case m > 0:
		return 1
	case m < 0:
		return -1
	default:
		return 0
	}
}

// Float64 возвращает сумму в рублях (или долларах) как float64.
// Только для отображения, не для вычислений!
func (m Money) Float64() float64 {
	return float64(m) / scaleF
}

// String форматирует сумму с двумя знаками после запятой.
func (m Money) String() string {
	if m == 0 {
		return "0.00"
	}
	negative := m < 0
	if negative {
		m = -m
	}
	units := int64(m) / scale
	cents := int64(m) % scale
	if negative {
		return fmt.Sprintf("-%d.%02d", units, cents)
	}
	return fmt.Sprintf("%d.%02d", units, cents)
}

// MarshalText реализует интерфейс encoding.TextMarshaler.
func (m Money) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText реализует интерфейс encoding.TextUnmarshaler.
func (m *Money) UnmarshalText(text []byte) error {
	val, err := ParseMoney(string(text))
	if err != nil {
		return err
	}
	*m = val
	return nil
}

// ──────────── Вспомогательные функции ────────────

// roundFloat округляет число с плавающей точкой до ближайшего целого.
func roundFloat(f float64) int64 {
	if f >= 0 {
		return int64(f + 0.5)
	}
	return int64(f - 0.5)
}
