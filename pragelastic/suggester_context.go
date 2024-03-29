package pragelastic

import "errors"

// SuggesterContextQuery is used to define context information within
// a suggestion request.
type ESSuggesterContextQuery interface {
	Source() (interface{}, error)
}

// ContextSuggester is a fast suggester for e.g. type-ahead completion that supports filtering and boosting based on contexts.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/suggester-context.html
// for more details.
type ESContextSuggester struct {
	ESSuggester
	name           string
	prefix         string
	field          string
	size           *int
	contextQueries []ESSuggesterContextQuery
}

// Creates a new context suggester.
func NewContextSuggester(name string) *ESContextSuggester {
	return &ESContextSuggester{
		name:           name,
		contextQueries: make([]ESSuggesterContextQuery, 0),
	}
}

func (q *ESContextSuggester) Name() string {
	return q.name
}

func (q *ESContextSuggester) Prefix(prefix string) *ESContextSuggester {
	q.prefix = prefix
	return q
}

func (q *ESContextSuggester) Field(field string) *ESContextSuggester {
	q.field = field
	return q
}

func (q *ESContextSuggester) Size(size int) *ESContextSuggester {
	q.size = &size
	return q
}

func (q *ESContextSuggester) ContextQuery(query ESSuggesterContextQuery) *ESContextSuggester {
	q.contextQueries = append(q.contextQueries, query)
	return q
}

func (q *ESContextSuggester) ContextQueries(queries ...ESSuggesterContextQuery) *ESContextSuggester {
	q.contextQueries = append(q.contextQueries, queries...)
	return q
}

// contextSuggesterRequest is necessary because the order in which
// the JSON elements are routed to Elasticsearch is relevant.
// We got into trouble when using plain maps because the text element
// needs to go before the completion element.
type contextSuggesterRequest struct {
	Prefix     string      `json:"prefix"`
	Completion interface{} `json:"completion"`
}

// Creates the source for the context suggester.
func (q *ESContextSuggester) Source(includeName bool) (interface{}, error) {
	cs := &contextSuggesterRequest{}

	if q.prefix != "" {
		cs.Prefix = q.prefix
	}

	suggester := make(map[string]interface{})
	cs.Completion = suggester

	if q.field != "" {
		suggester["field"] = q.field
	}
	if q.size != nil {
		suggester["size"] = *q.size
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

	if !includeName {
		return cs, nil
	}

	source := make(map[string]interface{})
	source[q.name] = cs
	return source, nil
}
