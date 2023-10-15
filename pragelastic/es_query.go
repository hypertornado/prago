package pragelastic

type ESQuery interface {
	Source() (interface{}, error)
}

type ESBoolQuery struct {
	ESQuery
	mustClauses    []ESQuery
	mustNotClauses []ESQuery
	filterClauses  []ESQuery
	shouldClauses  []ESQuery
	//boost              *float64
	//minimumShouldMatch string
	//adjustPureNegative *bool
	//queryName          string
}

func NewESBoolQuery() *ESBoolQuery {
	return &ESBoolQuery{
		mustClauses:    make([]ESQuery, 0),
		mustNotClauses: make([]ESQuery, 0),
		filterClauses:  make([]ESQuery, 0),
		shouldClauses:  make([]ESQuery, 0),
	}
}

func (q *ESBoolQuery) Must(queries ...ESQuery) *ESBoolQuery {
	q.mustClauses = append(q.mustClauses, queries...)
	return q
}

func (q *ESBoolQuery) MustNot(queries ...ESQuery) *ESBoolQuery {
	q.mustNotClauses = append(q.mustNotClauses, queries...)
	return q
}

func (q *ESBoolQuery) Filter(filters ...ESQuery) *ESBoolQuery {
	q.filterClauses = append(q.filterClauses, filters...)
	return q
}

func (q *ESBoolQuery) Should(queries ...ESQuery) *ESBoolQuery {
	q.shouldClauses = append(q.shouldClauses, queries...)
	return q
}

func (q *ESBoolQuery) Source() (interface{}, error) {
	// {
	//	"bool" : {
	//		"must" : {
	//			"term" : { "user" : "kimchy" }
	//		},
	//		"must_not" : {
	//			"range" : {
	//				"age" : { "from" : 10, "to" : 20 }
	//			}
	//		},
	//    "filter" : [
	//      ...
	//    ]
	//		"should" : [
	//			{
	//				"term" : { "tag" : "wow" }
	//			},
	//			{
	//				"term" : { "tag" : "elasticsearch" }
	//			}
	//		],
	//		"minimum_should_match" : 1,
	//		"boost" : 1.0
	//	}
	// }

	query := make(map[string]interface{})

	boolClause := make(map[string]interface{})
	query["bool"] = boolClause

	// must
	if len(q.mustClauses) == 1 {
		src, err := q.mustClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["must"] = src
	} else if len(q.mustClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.mustClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["must"] = clauses
	}

	// must_not
	if len(q.mustNotClauses) == 1 {
		src, err := q.mustNotClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["must_not"] = src
	} else if len(q.mustNotClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.mustNotClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["must_not"] = clauses
	}

	// filter
	if len(q.filterClauses) == 1 {
		src, err := q.filterClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["filter"] = src
	} else if len(q.filterClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.filterClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["filter"] = clauses
	}

	// should
	if len(q.shouldClauses) == 1 {
		src, err := q.shouldClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["should"] = src
	} else if len(q.shouldClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.shouldClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["should"] = clauses
	}

	/*if q.boost != nil {
		boolClause["boost"] = *q.boost
	}
	if q.minimumShouldMatch != "" {
		boolClause["minimum_should_match"] = q.minimumShouldMatch
	}
	if q.adjustPureNegative != nil {
		boolClause["adjust_pure_negative"] = *q.adjustPureNegative
	}
	if q.queryName != "" {
		boolClause["_name"] = q.queryName
	}*/

	return query, nil
}

type ESTermQuery struct {
	name  string
	value interface{}
	//boost     *float64
	//queryName string
}

func NewESTermQuery(name string, value interface{}) *ESTermQuery {
	return &ESTermQuery{name: name, value: value}
}

// Source returns JSON for the query.
func (q *ESTermQuery) Source() (interface{}, error) {
	source := make(map[string]interface{})
	tq := make(map[string]interface{})
	source["term"] = tq
	tq[q.name] = q.value
	return source, nil
}

type ESMatchQuery struct {
	name string
	text interface{}
}

func NewESMatchQuery(name string, text interface{}) *ESMatchQuery {
	return &ESMatchQuery{name: name, text: text}
}

func (q *ESMatchQuery) Source() (interface{}, error) {
	// {"match":{"name":{"query":"value","type":"boolean/phrase"}}}
	source := make(map[string]interface{})

	match := make(map[string]interface{})
	source["match"] = match

	query := make(map[string]interface{})
	match[q.name] = query
	query["query"] = q.text
	return source, nil
}

type ESTermsQuery struct {
	name   string
	values []interface{}
}

func NewESTermsQuery(name string, values ...interface{}) *ESTermsQuery {
	q := &ESTermsQuery{
		name:   name,
		values: make([]interface{}, 0),
	}
	if len(values) > 0 {
		q.values = append(q.values, values...)
	}
	return q
}

func (q *ESTermsQuery) Source() (interface{}, error) {
	source := make(map[string]interface{})
	params := make(map[string]interface{})
	source["terms"] = params

	params[q.name] = q.values

	return source, nil
}
