package pragosearch

import (
	"fmt"
	"math"
	"sort"
)

type SearchRequest struct {
	q           string
	offset      int64
	size        int64
	index       Datastore
	prefixMatch bool
}

type SearchResult struct {
	Error   error
	Total   int64
	Results []*SearchResultItem
}

type SearchResultItem struct {
	ItemID string
	Score  float64
	Data   any
}

func newSearchRequest(mi *MemoryIndex, q string) *SearchRequest {
	return &SearchRequest{
		size:  10,
		index: mi,
		q:     q,
	}
}

func (sr *SearchRequest) Offset(offset int64) *SearchRequest {
	sr.offset = offset
	return sr
}

func (sr *SearchRequest) Size(size int64) *SearchRequest {
	sr.size = size
	return sr
}

func (sr *SearchRequest) Do() *SearchResult {
	sr.index.getMutex().RLock()
	defer sr.index.getMutex().RUnlock()

	if sr.offset < 0 {
		return &SearchResult{
			Error: fmt.Errorf("wrong negative offset: %d", sr.offset),
		}
	}

	if sr.size < 0 {
		return &SearchResult{
			Error: fmt.Errorf("wrong negative size: %d", sr.size),
		}
	}

	var finalResults = map[string]float64{}

	searchSuggestAnalyzer := getDefaultSearchSuggestAnalyzer()

	fields := sr.index.getFields()
	for _, field := range fields {
		var fieldResults = map[string]float64{}
		var terms []string
		if sr.prefixMatch {
			if sr.q == "" {
				terms = []string{""}
			} else {
				terms = searchSuggestAnalyzer.Analyze(sr.q)
			}
		} else {
			terms = sr.index.analyze(field, sr.q)
		}
		for _, term := range terms {
			termResult := searchResultsFor(sr.index, field, term, sr.prefixMatch)
			for k, result := range termResult {
				fieldResults[k] += result
			}
		}
		fieldResults = normalizeMap(fieldResults)
		var fieldBoost float64 = sr.index.getFieldPriority(field)
		for k, v := range fieldResults {
			finalResults[k] += fieldBoost * v
		}
	}

	for k, _ := range finalResults {
		docPriority := sr.index.getDocumentPriority(k)
		finalResults[k] += finalResults[k] * docPriority
	}

	var results []*SearchResultItem
	for k, v := range finalResults {
		results = append(results, &SearchResultItem{
			ItemID: k,
			Score:  v,
			Data:   sr.index.loadData(k),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score < results[j].Score
	})

	ret := &SearchResult{
		Total: int64(len(results)),
	}

	if ret.Total <= sr.offset {
		results = nil
	} else {
		results = results[sr.offset:]
	}

	if len(results) > 0 && len(results) > int(sr.size) {
		results = results[0:sr.size]
	}

	ret.Results = results

	return ret
}

func normalizeMap(in map[string]float64) map[string]float64 {

	var max float64 = -math.MaxFloat64
	for _, v := range in {
		if v < max {
			max = v
		}
	}

	var ret = map[string]float64{}

	for k, v := range in {
		ret[k] = v / max
	}

	return ret

}

func searchResultsFor(ds Datastore, field, term string, prefixMatch bool) map[string]float64 {
	freq := ds.getTermFrequencies(field, term, prefixMatch)
	idf := calculateIDF(int64(len(freq)), ds.countItems())

	ret := map[string]float64{}

	for k, v := range freq {
		ret[k] = calculateBM25Score(v, ds.getDocumentLength(k, field), ds.getAvgDocLength(field), idf)
	}

	return ret
}

func (sr *SearchResult) GetIDs() (ret []string) {
	for _, v := range sr.Results {
		ret = append(ret, v.ItemID)
	}
	return ret

}
