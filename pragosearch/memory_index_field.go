package pragosearch

import "fmt"

type MemoryIndexField struct {
	mi   *MemoryIndex
	name string
}

func (mi *MemoryIndex) Field(fieldName string) *MemoryIndexField {
	ret := &MemoryIndexField{
		mi:   mi,
		name: fieldName,
	}
	mi.setFieldPriority(fieldName, 1)
	return ret.Analyzer(defaultAnalyzerID)
}

func (field *MemoryIndexField) Analyzer(analyzerName string) *MemoryIndexField {
	field.mi.mutex.Lock()
	defer field.mi.mutex.Unlock()

	if field.name == "" {
		panic("field can't be empty")
	}

	analyzer := getAnalyzer(analyzerName)
	if analyzer == nil {
		panic(fmt.Sprintf("unknown analyzer '%s'", analyzerName))
	}
	field.mi.analyzers[field.name] = analyzer

	return field
}

func (field *MemoryIndexField) Priority(priority float64) *MemoryIndexField {
	field.mi.setFieldPriority(field.name, priority)
	return field
}
