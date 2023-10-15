package pragelastic

type ESSorter interface {
	Source() (interface{}, error)
}

type ESSortInfo struct {
	ESSorter
	Field     string
	Ascending bool
}

func (info ESSortInfo) Source() (interface{}, error) {
	prop := make(map[string]interface{})
	if info.Ascending {
		prop["order"] = "asc"
	} else {
		prop["order"] = "desc"
	}
	source := make(map[string]interface{})
	source[info.Field] = prop
	return source, nil
}
