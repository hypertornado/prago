package pragosearch

const defaultAnalyzerID = "czech"
const defaultSuggestAnalyzerID = "czech_suggest"

func getDefaultSearchSuggestAnalyzer() *analyzer {
	return getAnalyzer(defaultSuggestAnalyzerID)
}

type analyzer struct {
	Name        string
	PreFilters  []func(string) string
	Tokenizer   func(string) []string
	PostFilters []func(string) string
}

var analyzers = []*analyzer{
	{
		Name: "czech",
		PreFilters: []func(string) string{
			lowercaser,
		},
		Tokenizer: tokenizer,
		PostFilters: []func(string) string{
			czechStemmer,
			isCzechStopword,
			removeDiacritics,
		},
	},
	{
		Name: "czech_suggest",
		PreFilters: []func(string) string{
			lowercaser,
		},
		Tokenizer:   tokenizer,
		PostFilters: []func(string) string{
			//czechStemmer,
			//isCzechStopword,
			//removeDiacritics,
		},
	},
}

func getAnalyzer(name string) *analyzer {
	for _, v := range analyzers {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (analyzer *analyzer) Analyze(str string) []string {

	for _, filter := range analyzer.PreFilters {
		str = filter(str)
	}

	tokens := analyzer.Tokenizer(str)

	for _, filter := range analyzer.PostFilters {
		tokens = useFilter(tokens, filter)
	}

	return tokens
}

func useFilter(arr []string, filter func(string) string) (ret []string) {
	for _, v := range arr {
		v = filter(v)
		if v != "" {
			ret = append(ret, v)
		}
	}
	return
}
