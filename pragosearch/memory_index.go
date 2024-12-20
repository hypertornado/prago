package pragosearch

import (
	"fmt"
	"sync"
)

type MemoryIndex struct {
	mutex *sync.RWMutex
	//defaultIndexAnalyzer *analyzer

	analyzers map[string]*analyzer

	documentPriority map[string]float64
	documentLength   map[[2]string]int64
	termFrequency    map[[3]string]int64
}

func NewMemoryIndex() *MemoryIndex {
	ret := &MemoryIndex{
		mutex:     &sync.RWMutex{},
		analyzers: map[string]*analyzer{},
		//defaultIndexAnalyzer: getAnalyzer(defaultAnalyzerID),
	}
	ret.DeleteAll()
	return ret
}

func (mi *MemoryIndex) SetAnalyzer(field, analyzerID string) error {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	if field == "" {
		return fmt.Errorf("field can't be empty")
	}

	if mi.analyzers[field] != nil {
		return fmt.Errorf("analyzer for field '%s' already set", field)
	}

	analyzer := getAnalyzer(analyzerID)
	if analyzer == nil {
		return fmt.Errorf("unknown analyzer '%s'", analyzerID)
	}

	mi.analyzers[field] = analyzer
	return nil
}

func (mi *MemoryIndex) Index(id string) *Indexer {
	return newIndexer(mi, id)
}

func (mi *MemoryIndex) indexItem(indexer *Indexer) error {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	if mi.containsItem(indexer.id) {
		err := mi.deleteItem(indexer.id)
		if err != nil {
			return fmt.Errorf("can't index item '%s': %s", indexer.id, err)
		}
	}

	mi.documentPriority[indexer.id] = indexer.priority
	for name, value := range indexer.fields {
		mi.addToIndex(indexer.id, name, value)
	}

	return nil
}

func (mi *MemoryIndex) DeleteAll() error {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	mi.documentPriority = map[string]float64{}
	mi.documentLength = map[[2]string]int64{}
	mi.termFrequency = map[[3]string]int64{}
	return nil
}

func (mi *MemoryIndex) addToIndex(itemID, fieldID, value string) {

	tokens := mi.analyzers[fieldID].Analyze(value)
	tokensMap := map[string]int64{}
	for _, v := range tokens {
		tokensMap[v] += 1
	}

	mi.documentLength[[2]string{itemID, fieldID}] = int64(len(tokens))

	for k, v := range tokensMap {
		mi.termFrequency[[3]string{itemID, fieldID, k}] = v
	}
}

func (mi *MemoryIndex) getMutex() *sync.RWMutex {
	return mi.mutex
}

func (mi *MemoryIndex) getDocumentPriority(id string) float64 {
	return mi.documentPriority[id]
}

func (mi *MemoryIndex) containsItem(id string) bool {
	_, ok := mi.documentPriority[id]
	return ok
}

func (mi *MemoryIndex) deleteItem(id string) error {
	if !mi.containsItem(id) {
		return fmt.Errorf("index does not contain item '%s'", id)
	}
	delete(mi.documentPriority, id)

	for k, _ := range mi.documentLength {
		if k[0] == id {
			delete(mi.documentLength, k)
		}
	}

	for k, _ := range mi.termFrequency {
		if k[0] == id {
			delete(mi.termFrequency, k)
		}
	}

	return nil
}

func (mi *MemoryIndex) countItems() int64 {
	return int64(len(mi.documentPriority))
}

func (mi *MemoryIndex) Query(q string) *SearchRequest {
	return newSearchRequest(mi, q)
}

func (mi *MemoryIndex) getDocumentLength(id, field string) int64 {
	return mi.documentLength[[2]string{id, field}]
}

func (mi *MemoryIndex) analyze(field, input string) []string {
	return mi.analyzers[field].Analyze(input)
	//return mi.defaultIndexAnalyzer.Analyze(input)
}

func (mi *MemoryIndex) getAvgDocLength(field string) float64 {
	var totalLength, totalCount int64
	for k, v := range mi.documentLength {
		if k[1] == field {
			totalLength += v
			totalCount += 1
		}
	}

	return float64(totalLength) / float64(totalCount)
}

func (mi *MemoryIndex) getFields() (ret []string) {
	var retMap = map[string]bool{}
	for k, _ := range mi.documentLength {
		retMap[k[1]] = true
	}

	for k := range retMap {
		ret = append(ret, k)
	}
	return
}

func (mi *MemoryIndex) getTermFrequencies(field, term string) map[string]int64 {
	freq := map[string]int64{}
	for k, v := range mi.termFrequency {
		if v <= 0 {
			continue
		}
		if k[1] != field {
			continue
		}
		if k[2] != term {
			continue
		}

		freq[k[0]] = v
	}
	return freq
}
