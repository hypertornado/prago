package pragobleve

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/blevesearch/bleve/v2"
	index "github.com/blevesearch/bleve_index_api"
)

type PragoBleveIndex[T any] struct {
	name   string
	pb     *PragoBleve
	bIndex bleve.Index
}

func NewIndex[T any](pb *PragoBleve) *PragoBleveIndex[T] {
	var val T
	name := reflect.TypeOf(val).Name()
	name = strings.ToLower(name)

	ret := PragoBleveIndex[T]{
		name: name,
		pb:   pb,
	}
	return &ret
}

func (index *PragoBleveIndex[T]) path() string {
	return index.pb.indexPath(index.name)
}

//https://github.com/blevesearch/bleve/blob/master/mapping/mapping_test.go

func (index *PragoBleveIndex[T]) Create() error {
	mapping := bleve.NewIndexMapping()

	//mapping.DefaultMapping.AddFieldMapping()

	bIndex, err := bleve.New(index.path(), mapping)
	if err != nil {
		return err
	}
	index.bIndex = bIndex
	return nil
}

const dataSaveField = "_data"

func (index *PragoBleveIndex[T]) Save(item *T) error {
	var itemVal = reflect.ValueOf(item).Elem()
	var itemType = reflect.TypeOf(*item)
	id := itemVal.FieldByName("ID").String()

	var dataMap = map[string]any{}

	for i := range itemVal.NumField() {
		fieldVal := itemVal.Field(i)
		fieldTyp := itemType.Field(i)
		dataMap[fieldTyp.Name] = fieldVal.Interface()
	}

	marshaledData, err := json.Marshal(item)
	must(err)
	dataMap[dataSaveField] = string(marshaledData)

	return index.bIndex.Index(id, dataMap)
}

func (index *PragoBleveIndex[T]) Size() int64 {
	count, err := index.bIndex.DocCount()
	must(err)
	return int64(count)
}

func (bindex *PragoBleveIndex[T]) Get(id string) *T {

	doc, err := bindex.bIndex.Document(id)
	if err != nil {
		return nil
	}

	var savedValue string
	doc.VisitFields(func(f index.Field) {
		if f.Name() == dataSaveField {
			savedValue = string(f.Value())
		}
	})

	var ret T
	err = json.Unmarshal([]byte(savedValue), &ret)
	if err != nil {
		return nil
	}

	return &ret
}
