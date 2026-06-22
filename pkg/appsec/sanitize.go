package appsec

import "html/template"

// SanitizeHTML экранирует специальные символы HTML (защита от XSS).
func SanitizeHTML(input string) string {
	return template.HTMLEscapeString(input)
}

// ValidID проверяет, что строка состоит только из букв, цифр, дефиса и подчёркивания.
func ValidID(id string) bool {
	if len(id) == 0 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// ValidNumber возвращает true, если строка содержит только цифры.
func ValidNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
