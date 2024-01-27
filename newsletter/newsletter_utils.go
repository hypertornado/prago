package newsletter

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/hypertornado/prago"

	stripmd "github.com/writeas/go-strip-markdown"
)

func mustGetSetting(id string) string {
	app := newsletters.app
	val, err := app.GetSetting(context.Background(), id)
	must(err)
	return val
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var csrfRandomness string

func RequestCSRF(request *prago.Request) string {
	if csrfRandomness == "" {
		csrfRandomness = RandomString(30)
	}
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", csrfRandomness, request.Request().UserAgent()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var seeded = false

// RandomString returns random string
func RandomString(n int) string {
	return RandomStringWithLetters(n, letters)
}

func RandomStringWithLetters(n int, letters []rune) string {
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

func unlocalized(in string) func(string) string {
	return func(string) string {
		return in
	}
}

func filterMarkdown(in string) string {
	in = stripmd.Strip(in)
	in = strings.Replace(in, "\n", " ", -1)
	return in
}

// CropMarkdown remove all markdown special characters
func cropMarkdown(text string, count int) string {
	text = filterMarkdown(text)
	return crop(text, count)
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
