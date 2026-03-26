package i18n

import "strings"

// LangInstruction returns the language instruction string for the Claude prompt.
func LangInstruction(code string) string {
	switch strings.ToLower(code) {
	case "pt", "pt-br":
		return "Respond in Brazilian Portuguese (pt-BR)."
	case "en", "en-us":
		return "Respond in English."
	case "es":
		return "Respond in Spanish."
	case "fr":
		return "Respond in French."
	case "de":
		return "Respond in German."
	case "it":
		return "Respond in Italian."
	case "ja":
		return "Respond in Japanese."
	case "ko":
		return "Respond in Korean."
	case "zh":
		return "Respond in Chinese (Simplified)."
	case "ru":
		return "Respond in Russian."
	default:
		if code == "" {
			return "Respond in Brazilian Portuguese (pt-BR)."
		}
		return "Respond in " + code + "."
	}
}
