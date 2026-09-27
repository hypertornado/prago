package prago

import "fmt"

type FormFilter struct {
	uuid           string
	filterFunction func(*listQuery) *listQuery
}

func (app *App) FormFilter(id string) *FormFilter {
	if app.formFilters == nil {
		app.formFilters = map[string]*FormFilter{}
	}

	uuid := pseudorandomString(id)
	if app.formFilters[uuid] != nil {
		panic(fmt.Sprintf("form filter %s already defined", id))
	}

	ret := &FormFilter{
		uuid: uuid,
		filterFunction: func(in *listQuery) *listQuery {
			return in
		},
	}

	app.formFilters[ret.uuid] = ret
	return ret
}

func (filter *FormFilter) Is(name string, value any) *FormFilter {
	oldFn := filter.filterFunction
	newFn := func(lq *listQuery) *listQuery {
		lq = oldFn(lq)
		lq = lq.Is(name, value)
		return lq
	}
	filter.filterFunction = newFn
	return filter
}

func (filter *FormFilter) Where(condition string, values ...any) *FormFilter {
	filter.filterFunction = func(lq *listQuery) *listQuery {
		return lq.where(condition, values...)
	}
	return filter
}
