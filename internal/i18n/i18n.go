package i18n

type Lang string

const (
	EN Lang = "en"
	RU Lang = "ru"
)

func T(lang Lang, key string) string {
	if m, ok := translations[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := translations[EN][key]; ok {
		return v
	}
	return key
}

func TFunc(lang Lang) func(string) string {
	return func(key string) string {
		return T(lang, key)
	}
}

func ParseLang(s string) Lang {
	switch s {
	case "ru":
		return RU
	default:
		return EN
	}
}
