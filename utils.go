package prago

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

func prettyFilename(s string) string {
	if len(s) > 100 {
		s = s[len(s)-99:]
	}
	items := strings.Split(s, ".")
	for i := range items {
		items[i] = prettyURL(items[i])
	}
	return strings.Join(items, ".")
}

func prettyURL(s string) string {
	return slug.Make(s)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var seeded = false

func randomString(n int) string {
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
func consoleQuestion(question string) bool {
	fmt.Printf("%s (yes|no)\n", question)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	if text == "yes\n" || text == "y\n" {
		return true
	}
	return false
}

func crop(in string, cropLength int) string {
	if cropLength < 0 {
		return in
	}
	if in == "" {
		return ""
	}
	inRune := []rune(in)
	if len(inRune) <= cropLength {
		return in
	} else {
		inRune = inRune[:cropLength]
		in = strings.TrimRightFunc(string(inRune), func(r rune) bool {
			if r == ' ' {
				return false
			} else {
				return true
			}
		})
		in = strings.TrimRightFunc(in, func(r rune) bool {
			if r == ' ' {
				return true
			} else {
				return false
			}
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

func humanizeNumber(i int64) (ret string) {
	return numberToString(int(i), ' ')
}

func humanizeFloat(i float64, locale string) string {
	ret := humanizeNumber(int64(i))

	defaultStr := fmt.Sprintf("%g", i)

	if strings.Contains(defaultStr, ".") {
		items := strings.Split(defaultStr, ".")
		switch locale {
		case "cs":
			ret += ","
		default:
			ret += "."
		}
		ret += items[1]
	}

	return ret
}

var monthsCS = []string{"Leden", "Únor", "Březen", "Duben", "Květen", "Červen", "Červenec", "Srpen", "Září", "Říjen", "Listopad", "Prosinec"}
var monthsEN = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

func monthName(order int64, locale string) string {
	if order < 1 || order > 12 {
		return ""
	}
	switch locale {
	case "cs":
		return monthsCS[order-1]
	default:
		return monthsEN[order-1]
	}
}
