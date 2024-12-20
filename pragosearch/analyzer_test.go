package pragosearch

import (
	"strings"
	"testing"
)

func TestAnalyzerCzech(t *testing.T) {

	analyzer := getAnalyzer("czech")

	for k, v := range [][2]string{
		{
			"Jeníček a Mařenk",
			"jenick;marenk",
		},
		{
			"panákové",
			"panako",
		},
	} {
		result := analyzer.Analyze(v[0])
		resultStr := strings.Join(result, ";")
		if resultStr != v[1] {
			t.Fatalf("%d) expected '%s', got '%s'", k, v[1], resultStr)
		}
	}

}
