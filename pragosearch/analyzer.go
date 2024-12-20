package pragosearch

const defaultAnalyzerID = "czech"

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
}

func getAnalyzer(name string) *analyzer {
	for _, v := range analyzers {
		if v.Name == name {
			return v
		}
	}
	return nil
}

type analyzer struct {
	Name        string
	PreFilters  []func(string) string
	Tokenizer   func(string) []string
	PostFilters []func(string) string
}

func (analyzer *analyzer) Analyze(input string) []string {

	for _, filter := range analyzer.PreFilters {
		input = filter(input)
	}

	tokens := analyzer.Tokenizer(input)

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
