package pragelastic

type ESSuggester interface {
	Name() string
	Source(includeName bool) (interface{}, error)
}
