package edit

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

func Start(text string) string {
	err := godotenv.Load("word.env")
	if err != nil {
		log.Printf("Failed to load .env file: %v", err)
	}

	wm := loadReplacements()

	finalText := replaceWords(text, wm)
	return finalText
}

func replaceWords(text string, replacements map[string]string) string {
	re := regexp.MustCompile(`(?i)[\p{L}\p{N}_]+`)

	return re.ReplaceAllStringFunc(text, func(word string) string {
		if replacement, ok := replacements[strings.ToLower(word)]; ok {
			return replacement
		}
		return word
	})
}

func loadReplacements() map[string]string {
	result := make(map[string]string)

	for _, pair := range strings.Split(os.Getenv("WORD_REPLACEMENTS"), ";") {
		oldWord, newWord, ok := strings.Cut(pair, ":")
		if !ok {
			continue
		}

		oldWord = strings.TrimSpace(oldWord)
		newWord = strings.TrimSpace(newWord)

		if oldWord != "" {
			result[oldWord] = newWord
		}
	}

	return result
}
