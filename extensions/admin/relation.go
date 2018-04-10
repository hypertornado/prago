package admin

type relation struct {
	resource *Resource
	field    string
}

func (r *Resource) AddRelation(r2 *Resource, field string) {
	r.relations = append(r.relations, relation{r2, field})
}
