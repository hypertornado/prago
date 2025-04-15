package pragosearch

import "sync"

type Datastore interface {
	indexItem(indexer *Indexer) error
	analyze(string, string) []string
	getFields() []string
	getTermFrequencies(field, term string, prefixMatch bool) map[string]int64
	getDocumentLength(id, field string) int64
	countItems() int64

	getAvgDocLength(field string) float64

	getMutex() *sync.RWMutex

	getDocumentPriority(id string) float64

	getFieldPriority(fieldName string) float64

	storeData(field string, data any)
	loadData(field string) any
}
