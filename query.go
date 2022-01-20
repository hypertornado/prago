package prago

type query struct {
	query    *listQuery
	db       dbIface
	isDebug  bool
	resource resourceIface
}

func (resource *Resource[T]) query() query {
	return query{
		query:    &listQuery{},
		db:       resource.getApp().db,
		resource: resource,
	}
}

func (q query) where(condition string, values ...interface{}) query {
	q.query.where(condition, values...)
	return q
}

func (q query) is(name string, value interface{}) query {
	return q.where(sqlFieldToQuery(name), value)
}

func (q query) debug() query {
	q.isDebug = true
	return q
}

func (q query) order(name string) query {
	q.query.addOrder(name, false)
	return q
}

func (q query) orderDesc(name string) query {
	q.query.addOrder(name, true)
	return q
}

func (q query) limit(limit int64) query {
	q.query.limit = limit
	return q
}

func (q query) offset(offset int64) query {
	q.query.offset = offset
	return q
}

func (q query) first() (interface{}, error) {
	var ret interface{}
	err := getFirstItem(q.resource, q.db, q.resource.getID(), &ret, q.query, q.isDebug)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (q query) list() (interface{}, error) {
	var items interface{}
	err := listItems(q.resource, q.db, q.resource.getID(), &items, q.query, q.isDebug)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (q query) count() (int64, error) {
	return countItems(q.db, q.resource.getID(), q.query, q.isDebug)
}

func (q query) deleteWithQuery(item interface{}) (int64, error) {
	return deleteItems(q.db, q.resource.getID(), q.query, q.isDebug)
}
