package pragelastic

type ESSearchSource struct {
	query                 ESQuery       // query
	postQuery             ESQuery       // post_filter
	sliceQuery            ESQuery       // slice
	from                  int           // from
	size                  int           // size
	explain               *bool         // explain
	version               *bool         // version
	seqNoAndPrimaryTerm   *bool         // seq_no_primary_term
	sorters               []ESSorter    // sort
	trackScores           *bool         // track_scores
	trackTotalHits        interface{}   // track_total_hits
	searchAfterSortValues []interface{} // search_after
	minScore              *float64      // min_score
	timeout               string        // timeout
	terminateAfter        *int          // terminate_after
	storedFieldNames      []string      // stored_fields
	//docvalueFields        DocvalueFields           // docvalue_fields
	//scriptFields          []*ScriptField           // script_fields
	//fetchSourceContext    *FetchSourceContext      // _source
	aggregations map[string]ESAggregation // aggregations / aggs
	//highlight             *Highlight               // highlight
	globalSuggestText string
	suggesters        []ESSuggester // suggest
	//rescores                 []*Rescore  // rescore
	defaultRescoreWindowSize *int
	indexBoosts              map[string]float64 // indices_boost
	stats                    []string           // stats
	//innerHits                map[string]*InnerHit
	//collapse                 *CollapseBuilder // collapse
	profile bool // profile
}

func NewESSearchSource() *ESSearchSource {
	return &ESSearchSource{
		from:         -1,
		size:         -1,
		aggregations: make(map[string]ESAggregation),
		//indexBoosts:  make(map[string]float64),
		//innerHits:    make(map[string]*InnerHit),
	}
}

func (s *ESSearchSource) Sort(field string, ascending bool) *ESSearchSource {
	s.sorters = append(s.sorters, ESSortInfo{Field: field, Ascending: ascending})
	return s
}

func (s *ESSearchSource) Aggregation(name string, aggregation ESAggregation) *ESSearchSource {
	s.aggregations[name] = aggregation
	return s
}

func (s *ESSearchSource) From(from int) *ESSearchSource {
	s.from = from
	return s
}

func (s *ESSearchSource) Size(size int) *ESSearchSource {
	s.size = size
	return s
}

func (s *ESSearchSource) Query(query ESQuery) *ESSearchSource {
	s.query = query
	return s
}

func (s *ESSearchSource) Source() (interface{}, error) {
	source := make(map[string]interface{})

	if s.from != -1 {
		source["from"] = s.from
	}
	if s.size != -1 {
		source["size"] = s.size
	}
	if s.timeout != "" {
		source["timeout"] = s.timeout
	}
	if s.terminateAfter != nil {
		source["terminate_after"] = *s.terminateAfter
	}
	if s.query != nil {
		src, err := s.query.Source()
		if err != nil {
			return nil, err
		}
		source["query"] = src
	}
	if s.postQuery != nil {
		src, err := s.postQuery.Source()
		if err != nil {
			return nil, err
		}
		source["post_filter"] = src
	}
	if s.minScore != nil {
		source["min_score"] = *s.minScore
	}
	if s.version != nil {
		source["version"] = *s.version
	}
	if s.explain != nil {
		source["explain"] = *s.explain
	}
	if s.profile {
		source["profile"] = s.profile
	}
	/*if s.fetchSourceContext != nil {
		src, err := s.fetchSourceContext.Source()
		if err != nil {
			return nil, err
		}
		source["_source"] = src
	}*/
	if s.storedFieldNames != nil {
		switch len(s.storedFieldNames) {
		case 1:
			source["stored_fields"] = s.storedFieldNames[0]
		default:
			source["stored_fields"] = s.storedFieldNames
		}
	}
	/*if len(s.docvalueFields) > 0 {
		src, err := s.docvalueFields.Source()
		if err != nil {
			return nil, err
		}
		source["docvalue_fields"] = src
	}*/
	/*if len(s.scriptFields) > 0 {
		sfmap := make(map[string]interface{})
		for _, scriptField := range s.scriptFields {
			src, err := scriptField.Source()
			if err != nil {
				return nil, err
			}
			sfmap[scriptField.FieldName] = src
		}
		source["script_fields"] = sfmap
	}*/
	if len(s.sorters) > 0 {
		var sortarr []interface{}
		for _, sorter := range s.sorters {
			src, err := sorter.Source()
			if err != nil {
				return nil, err
			}
			sortarr = append(sortarr, src)
		}
		source["sort"] = sortarr
	}
	if v := s.trackScores; v != nil {
		source["track_scores"] = *v
	}
	if v := s.trackTotalHits; v != nil {
		source["track_total_hits"] = v
	}
	if len(s.searchAfterSortValues) > 0 {
		source["search_after"] = s.searchAfterSortValues
	}
	if s.sliceQuery != nil {
		src, err := s.sliceQuery.Source()
		if err != nil {
			return nil, err
		}
		source["slice"] = src
	}
	if len(s.indexBoosts) > 0 {
		source["indices_boost"] = s.indexBoosts
	}
	if len(s.aggregations) > 0 {
		aggsMap := make(map[string]interface{})
		for name, aggregate := range s.aggregations {
			src, err := aggregate.Source()
			if err != nil {
				return nil, err
			}
			aggsMap[name] = src
		}
		source["aggregations"] = aggsMap
	}
	/*if s.highlight != nil {
		src, err := s.highlight.Source()
		if err != nil {
			return nil, err
		}
		source["highlight"] = src
	}*/
	if len(s.suggesters) > 0 {
		suggesters := make(map[string]interface{})
		for _, s := range s.suggesters {
			src, err := s.Source(false)
			if err != nil {
				return nil, err
			}
			suggesters[s.Name()] = src
		}
		if s.globalSuggestText != "" {
			suggesters["text"] = s.globalSuggestText
		}
		source["suggest"] = suggesters
	}
	/*if len(s.rescores) > 0 {
		// Strip empty rescores from request
		var rescores []*Rescore
		for _, r := range s.rescores {
			if !r.IsEmpty() {
				rescores = append(rescores, r)
			}
		}
		if len(rescores) == 1 {
			rescores[0].defaultRescoreWindowSize = s.defaultRescoreWindowSize
			src, err := rescores[0].Source()
			if err != nil {
				return nil, err
			}
			source["rescore"] = src
		} else {
			var slice []interface{}
			for _, r := range rescores {
				r.defaultRescoreWindowSize = s.defaultRescoreWindowSize
				src, err := r.Source()
				if err != nil {
					return nil, err
				}
				slice = append(slice, src)
			}
			source["rescore"] = slice
		}
	}*/
	if len(s.stats) > 0 {
		source["stats"] = s.stats
	}
	// TODO ext builders

	/*if s.collapse != nil {
		src, err := s.collapse.Source()
		if err != nil {
			return nil, err
		}
		source["collapse"] = src
	}*/

	if v := s.seqNoAndPrimaryTerm; v != nil {
		source["seq_no_primary_term"] = *v
	}

	/*if len(s.innerHits) > 0 {
		// Top-level inner hits
		// See http://www.elastic.co/guide/en/elasticsearch/reference/1.5/search-request-inner-hits.html#top-level-inner-hits
		// "inner_hits": {
		//   "<inner_hits_name>": {
		//     "<path|type>": {
		//       "<path-to-nested-object-field|child-or-parent-type>": {
		//         <inner_hits_body>,
		//         [,"inner_hits" : { [<sub_inner_hits>]+ } ]?
		//       }
		//     }
		//   },
		//   [,"<inner_hits_name_2>" : { ... } ]*
		// }
		m := make(map[string]interface{})
		for name, hit := range s.innerHits {
			if hit.path != "" {
				src, err := hit.Source()
				if err != nil {
					return nil, err
				}
				path := make(map[string]interface{})
				path[hit.path] = src
				m[name] = map[string]interface{}{
					"path": path,
				}
			} else if hit.typ != "" {
				src, err := hit.Source()
				if err != nil {
					return nil, err
				}
				typ := make(map[string]interface{})
				typ[hit.typ] = src
				m[name] = map[string]interface{}{
					"type": typ,
				}
			} else {
				// TODO the Java client throws here, because either path or typ must be specified
				_ = m
			}
		}
		source["inner_hits"] = m
	}*/

	return source, nil
}
