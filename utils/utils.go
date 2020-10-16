package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

//PrettyFilename converts filename to url-friendly form with regard to extension
func PrettyFilename(s string) string {
	if len(s) > 100 {
		s = s[len(s)-99:]
	}
	items := strings.Split(s, ".")
	for i := range items {
		items[i] = PrettyURL(items[i])
	}
	return strings.Join(items, ".")
}

//PrettyURL converts string to url-friendly form
func PrettyURL(s string) string {
	return slug.Make(s)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var seeded = false

//RandomString returns random string
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

//ConsoleQuestion asks for boolean answer in console
func ConsoleQuestion(question string) bool {
	fmt.Printf("%s (yes|no)\n", question)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if text == "yes\n" || text == "y\n" {
		return true
	}
	return false
}

//Crop removes text longer then count
//it tries not to split in the middle of words
func Crop(in string, cropLength int) string {
	if cropLength < 0 {
		return in
	}
	if in == "" {
		return ""
	}
	inRune := []rune(in)
	if len(inRune) < cropLength {
		return in
	} else {
		inRune = inRune[:cropLength]
		in = strings.TrimRightFunc(string(inRune), func(r rune) bool {
			if r == ' ' {
				return false
			}
			return true
		})
		in = strings.TrimRightFunc(in, func(r rune) bool {
			if r == ' ' {
				return true
			}
			return false
		})
		return in + "…"
	}
}

func numberToString(n int, sep rune) string {

	s := strconv.Itoa(n)

	startOffset := 0
	var buff bytes.Buffer

	if n < 0 {
		startOffset = 1
		buff.WriteByte('-')
	}

	l := len(s)

	commaIndex := 3 - ((l - startOffset) % 3)

	if commaIndex == 3 {
		commaIndex = 0
	}

	for i := startOffset; i < l; i++ {

		if commaIndex == 3 {
			buff.WriteRune(sep)
			commaIndex = 0
		}
		commaIndex++

		buff.WriteByte(s[i])
	}

	return buff.String()
}

func HumanizeNumber(i int64) (ret string) {
	return numberToString(int(i), ' ')
}

/*func HumanizeFloat(i float64, locale string) (ret string) {
	return numberToString(int(i), ' ')
}*/
