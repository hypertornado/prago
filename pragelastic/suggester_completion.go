package pragelastic

import "errors"

// CompletionSuggester is a fast suggester for e.g. type-ahead completion.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-suggesters-completion.html
// for more details.
type ESCompletionSuggester struct {
	ESSuggester
	name           string
	text           string
	prefix         string
	regex          string
	field          string
	analyzer       string
	size           *int
	shardSize      *int
	contextQueries []ESSuggesterContextQuery

	//fuzzyOptions   *FuzzyCompletionSuggesterOptions
	//regexOptions   *RegexCompletionSuggesterOptions
	skipDuplicates *bool
}

// Creates a new completion suggester.
func NewESCompletionSuggester(name string) *ESCompletionSuggester {
	return &ESCompletionSuggester{
		name: name,
	}
}

func (q *ESCompletionSuggester) Name() string {
	return q.name
}

func (q *ESCompletionSuggester) Text(text string) *ESCompletionSuggester {
	q.text = text
	return q
}

func (q *ESCompletionSuggester) Prefix(prefix string) *ESCompletionSuggester {
	q.prefix = prefix
	return q
}

func (q *ESCompletionSuggester) SkipDuplicates(skipDuplicates bool) *ESCompletionSuggester {
	q.skipDuplicates = &skipDuplicates
	return q
}

func (q *ESCompletionSuggester) Field(field string) *ESCompletionSuggester {
	q.field = field
	return q
}

func (q *ESCompletionSuggester) Analyzer(analyzer string) *ESCompletionSuggester {
	q.analyzer = analyzer
	return q
}

func (q *ESCompletionSuggester) Size(size int) *ESCompletionSuggester {
	q.size = &size
	return q
}

func (q *ESCompletionSuggester) ShardSize(shardSize int) *ESCompletionSuggester {
	q.shardSize = &shardSize
	return q
}

func (q *ESCompletionSuggester) ContextQuery(query ESSuggesterContextQuery) *ESCompletionSuggester {
	q.contextQueries = append(q.contextQueries, query)
	return q
}

func (q *ESCompletionSuggester) ContextQueries(queries ...ESSuggesterContextQuery) *ESCompletionSuggester {
	q.contextQueries = append(q.contextQueries, queries...)
	return q
}

// completionSuggesterRequest is necessary because the order in which
// the JSON elements are routed to Elasticsearch is relevant.
// We got into trouble when using plain maps because the text element
// needs to go before the completion element.
type completionSuggesterRequest struct {
	Text       string      `json:"text,omitempty"`
	Prefix     string      `json:"prefix,omitempty"`
	Regex      string      `json:"regex,omitempty"`
	Completion interface{} `json:"completion,omitempty"`
}

// Source creates the JSON data for the completion suggester.
func (q *ESCompletionSuggester) Source(includeName bool) (interface{}, error) {
	cs := &completionSuggesterRequest{}

	if q.text != "" {
		cs.Text = q.text
	}
	if q.prefix != "" {
		cs.Prefix = q.prefix
	}
	if q.regex != "" {
		cs.Regex = q.regex
	}

	suggester := make(map[string]interface{})
	cs.Completion = suggester

	if q.analyzer != "" {
		suggester["analyzer"] = q.analyzer
	}
	if q.field != "" {
		suggester["field"] = q.field
	}
	if q.size != nil {
		suggester["size"] = *q.size
	}
	if q.shardSize != nil {
		suggester["shard_size"] = *q.shardSize
	}
	switch len(q.contextQueries) {
	case 0:
	case 1:
		src, err := q.contextQueries[0].Source()
		if err != nil {
			return nil, err
		}
		suggester["contexts"] = src
	default:
		ctxq := make(map[string]interface{})
		for _, query := range q.contextQueries {
			src, err := query.Source()
			if err != nil {
				return nil, err
			}
			// Merge the dictionary into ctxq
			m, ok := src.(map[string]interface{})
			if !ok {
				return nil, errors.New("elastic: context query is not a map")
			}
			for k, v := range m {
				ctxq[k] = v
			}
		}
		suggester["contexts"] = ctxq
	}

	// Fuzzy options
	/*if q.fuzzyOptions != nil {
		src, err := q.fuzzyOptions.Source()
		if err != nil {
			return nil, err
		}
		suggester["fuzzy"] = src
	}

	// Regex options
	if q.regexOptions != nil {
		src, err := q.regexOptions.Source()
		if err != nil {
			return nil, err
		}
		suggester["regex"] = src
	}*/

	if q.skipDuplicates != nil {
		suggester["skip_duplicates"] = *q.skipDuplicates
	}

	// TODO(oe) Add completion-suggester specific parameters here

	if !includeName {
		return cs, nil
	}

	source := make(map[string]interface{})
	source[q.name] = cs
	return source, nil
}
