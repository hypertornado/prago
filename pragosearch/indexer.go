package pragosearch

import "fmt"

type Indexer struct {
	id       string
	priority float64
	fields   map[string]string
	ds       Datastore
}

func newIndexer(store Datastore, id string) *Indexer {
	ret := &Indexer{
		id:       id,
		priority: 1,
		fields:   map[string]string{},
		ds:       store,
	}
	return ret
}

func (i *Indexer) Priority(priority float64) *Indexer {
	i.priority = priority
	return i
}

func (i *Indexer) Set(field, value string) *Indexer {
	i.fields[field] = value
	return i
}

func (i *Indexer) StoreData(data any) *Indexer {
	i.ds.storeData(i.id, data)
	return i
}

func (i *Indexer) Do() error {
	if i.id == "" {
		return fmt.Errorf("no error set")
	}
	return i.ds.indexItem(i)
}
