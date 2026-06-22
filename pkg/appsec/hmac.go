package appsec

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
)

// SignURLParams добавляет к URL параметр `sig` с HMAC-SHA256 подписью.
// Параметры сортируются по ключу для стабильности.
func SignURLParams(baseURL string, params map[string]string, secret []byte) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramStr string
	for i, k := range keys {
		if i > 0 {
			paramStr += "&"
		}
		paramStr += k + "=" + params[k]
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(paramStr))
	sig := hex.EncodeToString(mac.Sum(nil))

	return baseURL + "?" + paramStr + "&sig=" + sig, nil
}

// VerifyURLParams проверяет подпись `sig` в URL. Возвращает ошибку, если подпись не совпадает.
func VerifyURLParams(fullURL string, secret []byte) error {
	u, err := url.Parse(fullURL)
	if err != nil {
		return err
	}
	q := u.Query()
	sig := q.Get("sig")
	if sig == "" {
		return errors.New("missing signature")
	}
	// Удаляем sig из набора для проверки
	delete(q, "sig")
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramStr string
	for i, k := range keys {
		if i > 0 {
			paramStr += "&"
		}
		paramStr += k + "=" + q.Get(k)
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(paramStr))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return errors.New("invalid signature")
	}
	return nil
}
