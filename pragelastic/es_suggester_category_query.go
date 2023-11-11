package pragelastic

type ESSuggesterCategoryQuery struct {
	name   string
	values map[string]*int
}

// NewSuggesterCategoryQuery creates a new SuggesterCategoryQuery.
func NewESSuggesterCategoryQuery(name string, values ...string) *ESSuggesterCategoryQuery {
	q := &ESSuggesterCategoryQuery{
		name:   name,
		values: make(map[string]*int),
	}

	if len(values) > 0 {
		q.Values(values...)
	}
	return q
}

func (q *ESSuggesterCategoryQuery) Value(val string) *ESSuggesterCategoryQuery {
	q.values[val] = nil
	return q
}

func (q *ESSuggesterCategoryQuery) ValueWithBoost(val string, boost int) *ESSuggesterCategoryQuery {
	q.values[val] = &boost
	return q
}

func (q *ESSuggesterCategoryQuery) Values(values ...string) *ESSuggesterCategoryQuery {
	for _, val := range values {
		q.values[val] = nil
	}
	return q
}

// Source returns a map that will be used to serialize the context query as JSON.
func (q *ESSuggesterCategoryQuery) Source() (interface{}, error) {
	source := make(map[string]interface{})

	switch len(q.values) {
	case 0:
		source[q.name] = make([]string, 0)
	default:
		contexts := make([]interface{}, 0)
		for val, boost := range q.values {
			context := make(map[string]interface{})
			context["context"] = val
			if boost != nil {
				context["boost"] = *boost
			}
			contexts = append(contexts, context)
		}
		source[q.name] = contexts
	}

	return source, nil
}
