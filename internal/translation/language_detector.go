package translation

import (
	"strings"
	"sync"

	"github.com/abadojack/whatlanggo"
)

// LanguageDetector handles language detection using whatlanggo
// whatlanggo is a pure Go implementation with minimal binary size impact
type LanguageDetector struct {
	once sync.Once
}

// languageDetectorInstance is the singleton instance
var (
	languageDetectorInstance *LanguageDetector
	languageDetectorOnce     sync.Once
)

// GetLanguageDetector returns the singleton language detector instance
func GetLanguageDetector() *LanguageDetector {
	languageDetectorOnce.Do(func() {
		languageDetectorInstance = &LanguageDetector{}
	})
	return languageDetectorInstance
}

// DetectLanguage detects the language of the given text
// Returns the ISO 639-1 language code (e.g., "en", "zh", "ja")
// Returns empty string if detection fails or confidence is too low
func (ld *LanguageDetector) DetectLanguage(text string) string {
	if text == "" {
		return ""
	}

	// Clean text for better detection
	text = strings.TrimSpace(text)
	if len(text) < 3 {
		return ""
	}

	// Remove HTML tags if present
	cleanText := removeHTMLTags(text)
	textForDetection := text

	// Only use cleaned text if it's significantly different and has enough content
	if len(cleanText) > 10 && len(cleanText) < len(text) {
		textForDetection = cleanText
	}

	// Detect language with options
	// Use whitelist to only detect languages we support
	supportedLangs := supportedLanguages()
	whitelist := make(map[whatlanggo.Lang]bool)
	for _, lang := range supportedLangs {
		whitelist[lang] = true
	}

	options := whatlanggo.Options{
		Whitelist: whitelist,
	}

	info := whatlanggo.DetectWithOptions(textForDetection, options)

	// Check confidence level - only accept high confidence detections
	if info.Confidence < 0.5 {
		return ""
	}

	// Convert whatlanggo Lang to ISO 639-1 code
	detectedCode := whatlangToISOCode(info.Lang)
	return detectedCode
}

// ShouldTranslate determines if translation is needed based on language detection
// Returns true if:
// - Language detection fails (fallback to translation for safety)
// - Detected language differs from target language
// Returns false if text is already in target language
func (ld *LanguageDetector) ShouldTranslate(text, targetLang string) bool {
	detectedLang := ld.DetectLanguage(text)

	// If detection failed, assume translation is needed (fallback behavior)
	if detectedLang == "" {
		return true
	}

	// Normalize language codes for comparison
	detectedLang = normalizeLangCode(detectedLang)
	targetLang = normalizeLangCode(targetLang)

	// Check if already in target language
	return detectedLang != targetLang
}

// supportedLanguages returns the list of languages we want to detect
func supportedLanguages() []whatlanggo.Lang {
	return []whatlanggo.Lang{
		whatlanggo.Eng,
		whatlanggo.Cmn, // Chinese (Mandarin)
		whatlanggo.Jpn,
		whatlanggo.Kor,
		whatlanggo.Spa,
		whatlanggo.Fra,
		whatlanggo.Deu,
		whatlanggo.Por,
		whatlanggo.Rus,
		whatlanggo.Ita,
		whatlanggo.Nld,
		whatlanggo.Pol,
		whatlanggo.Tur,
		whatlanggo.Vie,
		whatlanggo.Tha,
		whatlanggo.Ind,
		whatlanggo.Hin,
	}
}

// whatlangToISOCode converts whatlanggo Lang to ISO 639-1 code
func whatlangToISOCode(lang whatlanggo.Lang) string {
	langMap := map[whatlanggo.Lang]string{
		whatlanggo.Eng: "en",
		whatlanggo.Cmn: "zh",
		whatlanggo.Jpn: "ja",
		whatlanggo.Kor: "ko",
		whatlanggo.Spa: "es",
		whatlanggo.Fra: "fr",
		whatlanggo.Deu: "de",
		whatlanggo.Por: "pt",
		whatlanggo.Rus: "ru",
		whatlanggo.Ita: "it",
		whatlanggo.Nld: "nl",
		whatlanggo.Pol: "pl",
		whatlanggo.Tur: "tr",
		whatlanggo.Vie: "vi",
		whatlanggo.Tha: "th",
		whatlanggo.Ind: "id",
		whatlanggo.Hin: "hi",
	}

	if code, ok := langMap[lang]; ok {
		return code
	}
	return ""
}

// normalizeLangCode normalizes language codes (e.g., "zh-CN" -> "zh", "en-US" -> "en")
func normalizeLangCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if len(code) > 2 {
		code = code[:2]
	}
	return code
}

// removeHTMLTags removes HTML tags from text for better language detection
func removeHTMLTags(text string) string {
	// Simple HTML tag removal
	var result strings.Builder
	inTag := false
	for _, r := range text {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}
