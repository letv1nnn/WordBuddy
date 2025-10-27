package apirequests

import (
	"os"
	"strings"
	"sync"

	"github.com/translated/lara-go/lara"
)

func TranslateText(text string, langFrom string, langTo string) ([]string, error) {
	LARA_ACCESS_KEY_ID := os.Getenv("TRANSLATE_ID")
	LARA_ACCESS_KEY_SECRET := os.Getenv("TRANSLATE_SECRET")
	credentials := lara.NewCredentials(LARA_ACCESS_KEY_ID, LARA_ACCESS_KEY_SECRET)
	lara_translator := lara.NewTranslator(credentials, nil)

	words := strings.Split(text, " ")
	translated := make([]string, len(words))

	var wg sync.WaitGroup
	for i, word := range words {
		wg.Add(1)
		go func(i int, word string) {
			defer wg.Done()
			translated[i] = translateWord(lara_translator, word, langFrom, langTo)
		}(i, word)
	}
	wg.Wait()

	return translated, nil
}

func translateWord(laraTranslator *lara.Translator, word string, languageFrom string, languageTo string) string {
	languageFrom = convertLangFormat(strings.ToLower(languageFrom))
	languageTo = convertLangFormat(strings.ToLower(languageTo))

	result, err := laraTranslator.Translate(
		word,
		languageFrom,
		languageTo,
		lara.TranslateOptions{},
	)
	if err != nil {
		return "failed to translate"
	}
	return *result.Translation.String
}

func convertLangFormat(lang string) string {
	langMap := map[string]string{
		"english":    "en-US",
		"spanish":    "es-ES",
		"french":     "fr-FR",
		"german":     "de-DE",
		"russian":    "ru-RU",
		"chinese":    "zh-CN",
		"japanese":   "ja-JP",
		"korean":     "ko-KR",
		"italian":    "it-IT",
		"portuguese": "pt-PT",
		"arabic":     "ar-SA",
		"hindi":      "hi-IN",
		"bengali":    "bn-BD",
		"turkish":    "tr-TR",
		"dutch":      "nl-NL",
		"polish":     "pl-PL",
		"swedish":    "sv-SE",
		"norwegian":  "no-NO",
		"finnish":    "fi-FI",
	}

	if code, exists := langMap[lang]; exists {
		return code
	}
	return "unknown"
}
