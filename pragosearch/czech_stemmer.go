package pragosearch

import (
	_ "embed"
	"strings"
)

func czechStemmer(word string) string {
	word = czechRemoveCase(word)
	word = czechRemovePossessives(word)
	if len(word) > 0 {
		word = czechNormalize(word)
	}
	return word

}

func czechRemoveCase(word string) string {
	length := len(word)

	// Remove 5 chars if ends with "atech" and length > 7
	if length > 7 && strings.HasSuffix(word, "atech") {
		return word[:length-5] // equivalent to word[0..-6] in Ruby
	}

	// Remove 4 chars if ends with certain suffixes and length > 6
	if length > 6 && (strings.HasSuffix(word, "ětem") ||
		strings.HasSuffix(word, "etem") ||
		strings.HasSuffix(word, "atům")) {
		return word[:length-4] // equivalent to word[0..-5]
	}

	// Remove 3 chars if ends with any of these suffixes and length > 5
	if length > 5 && (strings.HasSuffix(word, "ech") ||
		strings.HasSuffix(word, "ich") ||
		strings.HasSuffix(word, "ích") ||
		strings.HasSuffix(word, "ého") ||
		strings.HasSuffix(word, "ěmi") ||
		strings.HasSuffix(word, "emi") ||
		strings.HasSuffix(word, "ému") ||
		strings.HasSuffix(word, "ěte") ||
		strings.HasSuffix(word, "ete") ||
		strings.HasSuffix(word, "ěti") ||
		strings.HasSuffix(word, "eti") ||
		strings.HasSuffix(word, "ího") ||
		strings.HasSuffix(word, "iho") ||
		strings.HasSuffix(word, "ími") ||
		strings.HasSuffix(word, "ímu") ||
		strings.HasSuffix(word, "imu") ||
		strings.HasSuffix(word, "ách") ||
		strings.HasSuffix(word, "ata") ||
		strings.HasSuffix(word, "aty") ||
		strings.HasSuffix(word, "ých") ||
		strings.HasSuffix(word, "ama") ||
		strings.HasSuffix(word, "ami") ||
		strings.HasSuffix(word, "ové") ||
		strings.HasSuffix(word, "ovi") ||
		strings.HasSuffix(word, "ými")) {
		return word[:length-3] // equivalent to word[0..-4]
	}

	// Remove 2 chars if ends with any of these suffixes and length > 4
	if length > 4 && (strings.HasSuffix(word, "em") ||
		strings.HasSuffix(word, "es") ||
		strings.HasSuffix(word, "ém") ||
		strings.HasSuffix(word, "ím") ||
		strings.HasSuffix(word, "ům") ||
		strings.HasSuffix(word, "at") ||
		strings.HasSuffix(word, "ám") ||
		strings.HasSuffix(word, "os") ||
		strings.HasSuffix(word, "us") ||
		strings.HasSuffix(word, "ým") ||
		strings.HasSuffix(word, "mi") ||
		strings.HasSuffix(word, "ou")) {
		return word[:length-2] // equivalent to word[0..-3]
	}

	// Remove 1 char if ends with certain vowels and length > 3
	vowels := []string{"a", "e", "i", "o", "u", "ů", "y", "á", "é", "í", "ý", "ě"}
	if length > 3 {
		lastChar := word[length-1:]
		for _, v := range vowels {
			if lastChar == v {
				return word[:length-1] // equivalent to word[0..-2]
			}
		}
	}

	return word
}

func czechRemovePossessives(word string) string {
	length := len(word)
	// Remove 2 chars if ends with "ov", "in", or "ův" and length > 5
	if length > 5 && (strings.HasSuffix(word, "ov") ||
		strings.HasSuffix(word, "in") ||
		strings.HasSuffix(word, "ův")) {
		return word[:length-2] // equivalent to word[0..-3]
	}
	return word
}

func czechNormalize(word string) string {
	length := len(word)

	if strings.HasSuffix(word, "čt") {
		return word[:length-2] + "ck"
	}
	if strings.HasSuffix(word, "št") {
		return word[:length-2] + "sk"
	}
	if strings.HasSuffix(word, "c") {
		return word[:length-1] + "k"
	}
	if strings.HasSuffix(word, "č") {
		return word[:length-1] + "k"
	}
	if strings.HasSuffix(word, "z") {
		return word[:length-1] + "h"
	}
	if strings.HasSuffix(word, "ž") {
		return word[:length-1] + "h"
	}

	// If the second-last char is 'e'
	if length > 1 && word[length-2:length-1] == "e" {
		lastChar := word[length-1:]
		return word[:length-2] + lastChar
	}

	// If the second-last char is 'ů'
	if length > 2 && word[length-2:length-1] == "ů" {
		lastChar := word[length-1:]
		return word[:length-2] + "o" + lastChar
	}

	return word
}

//go:embed czech-stopwords.txt
var czechStopwordsTxt string

var czechStopwords = map[string]bool{}

func init() {
	words := strings.Split(czechStopwordsTxt, "\n")
	for _, v := range words {
		czechStopwords[v] = true
	}
}

func isCzechStopword(word string) string {
	if czechStopwords[word] {
		return ""
	}
	return word
}
