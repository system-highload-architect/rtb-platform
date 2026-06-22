package zerocopy

import (
	"strconv"
	"sync"
	"unsafe"
)

// ──────────────────────────── Байтовый пул ────────────────────────────

// bufferPool — пул для временных []byte буферов.
// Используется при сборке ответов, логировании и т.д.
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

// GetBytes возвращает *[]byte из пула с нулевой длиной и capacity ≥ 4096.
// После использования буфер нужно вернуть через PutBytes.
func GetBytes() *[]byte {
	return bufferPool.Get().(*[]byte)
}

// PutBytes обнуляет длину буфера и возвращает его в пул.
func PutBytes(buf *[]byte) {
	*buf = (*buf)[:0]
	bufferPool.Put(buf)
}

// ────────────── Zero‑copy конвертация string ↔ []byte ──────────────

// StringToBytes возвращает []byte, указывающий на те же данные, что и s.
// ⚠️ Изменять возвращаемый слайс НЕЛЬЗЯ — это нарушит иммутабельность строки.
// Безопасно, если слайс используется только для чтения.
func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString возвращает строку, указывающую на данные b.
// ⚠️ Нельзя изменять b после вызова, пока строка используется.
// Безопасно, если b не модифицируется.
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// ───────────── Инлайн‑парсинг JSON без аллокаций ─────────────

// GetJSONField извлекает необработанное значение поля field из JSON-объекта data.
// Возвращает слайс, содержащий JSON-значение (строку в кавычках, число, объект и т.д.).
// data должен быть валидным JSON-объектом. Аллокаций нет: возвращается подсегмент data.
// Если поле не найдено, возвращает nil, false.
func GetJSONField(data []byte, field string) ([]byte, bool) {
	// Используем json.Decoder, но без аллокаций: он работает поверх []byte.
	// Чтобы избежать аллокаций на ключах и значениях, мы воспользуемся
	// простым сканером, который ищет ключ и копирует диапазон значения.
	// Однако json.Decoder всё равно создаёт строки. Для настоящего zero‑alloc
	// нужен собственный лексер. При этом для простоты мы можем использовать
	// json.RawMessage и json.Token, но это приведёт к аллокациям.
	// Поэтому реализуем наивный, но быстрый поиск по байтам.

	// Будем искать подстроку `"field":` и затем извлекаем следующее JSON-значение.
	// Это не полностью устойчиво к пробелам, но для контролируемого входа из RTB подходит.
	// Для надёжности можно использовать библиотеку jsonparser (github.com/buger/jsonparser),
	// которая работает без аллокаций. Но мы не будем тащить внешнюю зависимость ради
	// одного простого кейса. Реализуем облегчённую версию.

	// TODO: заменить на более надёжный сканер при необходимости.
	return getJSONFieldSimple(data, field)
}

func getJSONFieldSimple(data []byte, field string) ([]byte, bool) {
	// Ищем ключ: "field"
	key := []byte(`"` + field + `"`)
	pos := indexOf(data, key)
	if pos == -1 {
		return nil, false
	}
	rest := data[pos+len(key):]
	// Пропускаем пробелы и двоеточие
	i := 0
	for i < len(rest) {
		switch rest[i] {
		case ' ', '\t', '\n', '\r':
			i++
		case ':':
			i++
			// Пропускаем пробелы после двоеточия
			for i < len(rest) && isSpace(rest[i]) {
				i++
			}
			valueEnd := findJSONValueEnd(rest, i)
			if valueEnd == -1 {
				return nil, false
			}
			return rest[i:valueEnd], true
		default:
			return nil, false // невалидный JSON
		}
	}
	return nil, false
}

func indexOf(data, pat []byte) int {
	for i := 0; i <= len(data)-len(pat); i++ {
		match := true
		for j := 0; j < len(pat); j++ {
			if data[i+j] != pat[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// findJSONValueEnd находит конец следующего JSON-значения, начиная с start.
// Возвращает индекс за последним символом значения или -1.
func findJSONValueEnd(data []byte, start int) int {
	if start >= len(data) {
		return -1
	}
	// Определяем тип значения по первому символу
	switch data[start] {
	case '"':
		// строка, ищем закрывающую кавычку с учётом экранирования
		for i := start + 1; i < len(data); i++ {
			if data[i] == '\\' {
				i++ // пропустить экранированный символ
				continue
			}
			if data[i] == '"' {
				return i + 1
			}
		}
		return -1
	case 't', 'f':
		// true/false
		for i := start; i < len(data); i++ {
			if isDelimiter(data[i]) {
				return i
			}
		}
		return len(data) // значение в конце данных
	case 'n':
		// null
		for i := start; i < len(data); i++ {
			if isDelimiter(data[i]) {
				return i
			}
		}
		return len(data)
	case '{', '[':
		// объект или массив — ищем закрывающую скобку с учётом вложенности
		return findBracketEnd(data, start)
	default:
		// число (включая отрицательные, дробные, экспоненциальные)
		for i := start + 1; i < len(data); i++ {
			if isDelimiter(data[i]) {
				return i
			}
		}
		return len(data)
	}
}

func isDelimiter(b byte) bool {
	return b == ',' || b == '}' || b == ']' || isSpace(b) || b == 0
}

func findBracketEnd(data []byte, start int) int {
	open := data[start]
	close := byte('}')
	if open == '[' {
		close = ']'
	}
	depth := 1
	inString := false
	for i := start + 1; i < len(data); i++ {
		if inString {
			if data[i] == '\\' {
				i++
				continue
			}
			if data[i] == '"' {
				inString = false
			}
			continue
		}
		switch data[i] {
		case '"':
			inString = true
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

// ───────────────── Утилиты для построения JSON ─────────────────

// AppendJSONString добавляет строку s в виде JSON-строки (с экранированием) к dst.
// Возвращает расширенный слайс. Аллокаций не делает, если capacity позволяет.
func AppendJSONString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '"', '\\':
			dst = append(dst, '\\', c)
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			dst = append(dst, c)
		}
	}
	dst = append(dst, '"')
	return dst
}

// AppendJSONInt добавляет целое число в JSON к dst (без аллокаций).
func AppendJSONInt(dst []byte, n int64) []byte {
	return strconv.AppendInt(dst, n, 10)
}
