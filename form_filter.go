package prago

type FormFilter struct {
	uuid           string
	filterFunction func(*listQuery) *listQuery
}

func (app *App) FormFilter() *FormFilter {

	if app.formFilters == nil {
		app.formFilters = map[string]*FormFilter{}
	}

	ret := &FormFilter{
		uuid: randomString(30),
		filterFunction: func(in *listQuery) *listQuery {
			return in
		},
	}

	app.formFilters[ret.uuid] = ret
	return ret
}

func (filter *FormFilter) Is(name string, value interface{}) *FormFilter {
	oldFn := filter.filterFunction
	newFn := func(lq *listQuery) *listQuery {
		lq = oldFn(lq)
		lq = lq.Is(name, value)
		return lq
	}
	filter.filterFunction = newFn
	return filter
}

func (filter *FormFilter) Where(condition string, values ...interface{}) *FormFilter {
	filter.filterFunction = func(lq *listQuery) *listQuery {
		lq = filter.filterFunction(lq)
		return lq.where(condition, values...)
	}
	return filter
}
