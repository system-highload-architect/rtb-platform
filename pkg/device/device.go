package device

import "unicode"

// DeviceType — тип устройства.
type DeviceType int

const (
	UnknownDevice DeviceType = iota
	Desktop
	Mobile
	Tablet
	Bot
)

// DeviceInfo — результат парсинга User-Agent.
type DeviceInfo struct {
	Type    DeviceType
	OS      string
	Browser string
}

// Parse извлекает информацию об устройстве из User-Agent строки.
// Возвращаемые строки — статические константы (без ссылок на исходный ua).
// Аллокаций не делает.
func Parse(ua string) DeviceInfo {
	return DeviceInfo{
		Type:    detectDeviceType(ua),
		OS:      detectOS(ua),
		Browser: detectBrowser(ua),
	}
}

// ── Определение типа устройства ──

func detectDeviceType(ua string) DeviceType {
	if containsIgnoreCase(ua, "bot") ||
		containsIgnoreCase(ua, "crawler") ||
		containsIgnoreCase(ua, "spider") ||
		containsIgnoreCase(ua, "scraper") {
		return Bot
	}
	if containsIgnoreCase(ua, "tablet") ||
		containsIgnoreCase(ua, "ipad") ||
		containsIgnoreCase(ua, "android") && !containsIgnoreCase(ua, "mobile") {
		return Tablet
	}
	if containsIgnoreCase(ua, "mobile") ||
		containsIgnoreCase(ua, "iphone") ||
		containsIgnoreCase(ua, "ipod") ||
		containsIgnoreCase(ua, "android") ||
		containsIgnoreCase(ua, "blackberry") ||
		containsIgnoreCase(ua, "windows phone") {
		return Mobile
	}
	return Desktop
}

// ── Определение ОС ──

var osKeywords = []struct {
	keyword string
	name    string
}{
	{"windows phone", "Windows Phone"},
	{"windows nt", "Windows"},
	{"iphone", "iOS"},
	{"ipad", "iOS"},
	{"ipod", "iOS"},
	{"mac os x", "macOS"},
	{"macintosh", "macOS"},
	{"android", "Android"},
	{"cros", "ChromeOS"},
	{"linux", "Linux"},
}

func detectOS(ua string) string {
	for _, o := range osKeywords {
		if containsIgnoreCase(ua, o.keyword) {
			return o.name
		}
	}
	return "Other"
}

// ── Определение браузера ──

var browserKeywords = []struct {
	keyword string
	name    string
}{
	{"edg/", "Edge"},
	{"edge/", "Edge"},
	{"opr/", "Opera"},
	{"opera mini", "Opera Mini"},
	{"chrome/", "Chrome"},
	{"safari/", "Safari"},
	{"firefox/", "Firefox"},
	{"msie ", "IE"},
	{"trident/", "IE"},
}

func detectBrowser(ua string) string {
	for _, b := range browserKeywords {
		if containsIgnoreCase(ua, b.keyword) {
			return b.name
		}
	}
	return "Other"
}

// ── Вспомогательная: поиск подстроки без учёта регистра и без аллокаций ──

func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			r1 := unicode.ToLower(rune(s[i+j]))
			r2 := unicode.ToLower(rune(substr[j]))
			if r1 != r2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
