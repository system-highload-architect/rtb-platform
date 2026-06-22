package appsec

import (
	"fmt"
	"net/url"
	"strings"
)

// SafeRedirect проверяет, что URL ведёт на разрешённый хост.
// Если URL относительный, он считается безопасным.
func SafeRedirect(rawURL string, allowedHosts []string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// Относительные URL (например, "/dashboard") не требуют проверки хоста
	if u.Host == "" {
		return rawURL, nil
	}
	host := strings.ToLower(u.Hostname())
	for _, allowed := range allowedHosts {
		if strings.EqualFold(host, allowed) {
			return rawURL, nil
		}
	}
	return "", fmt.Errorf("redirect to untrusted host: %s", host)
}
