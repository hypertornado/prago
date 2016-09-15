package utils

import (
	"bufio"
	"fmt"
	"github.com/Machiel/slugify"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

func PrettyFilename(s string) string {
	if len(s) > 100 {
		s = s[len(s)-99:]
	}
	items := strings.Split(s, ".")
	for i, _ := range items {
		items[i] = PrettyUrl(items[i])
	}
	return strings.Join(items, ".")
}

func PrettyUrl(s string) string {
	return slugify.Slugify(s)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var seeded = false

func RandomString(n int) string {
	if !seeded {
		rand.Seed(time.Now().Unix())
		seeded = true
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func ConsoleQuestion(question string) bool {
	fmt.Printf("%s (yes|no)\n", question)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if text == "yes\n" || text == "y\n" {
		return true
	}
	return false
}

func ColumnName(fieldName string) string {
	return PrettyUrl(fieldName)
}

func filterMarkdown(in string) string {
	r := regexp.MustCompile("\\[([^\\]]+)\\]\\(([^)]+)\\)")
	return r.ReplaceAllString(in, "$1")
}

//TODO: now just links are filtered, better will be to use markdown custom renderers
//https://godoc.org/github.com/russross/blackfriday#Renderer
func CropMarkdown(text string, count int) string {
	text = filterMarkdown(text)
	return Crop(text, count)
}

func Crop(text string, count int) string {
	runes := []rune(text)
	if len(runes) <= count {
		return text
	} else {
		ret := string(runes[0:count])
		i := strings.LastIndex(ret, " ")
		if i < 0 {
			return text
		}
		return ret[0:i]
	}
}
