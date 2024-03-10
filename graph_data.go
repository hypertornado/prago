package prago

import "sort"

type graphDataSource interface {
	Len() int
	Name(int) string
	Value(int) float64
}

func (graph *graph) View() *graphDataView {
	ret := &graphDataView{
		Name: graph.Name,
	}

	for i := 0; i < graph.dataSource.Len(); i++ {
		item := &graphDataViewItem{
			Name:  graph.dataSource.Name(i),
			Value: graph.dataSource.Value(i),
		}
		ret.Items = append(ret.Items, item)
	}

	var max float64
	for k, v := range ret.Items {
		if k == 0 || max < v.Value {
			max = v.Value
		}
	}

	for _, v := range ret.Items {
		v.Percent = (v.Value * 100) / max
	}

	return ret

}

type graphDataView struct {
	Name  string
	Items []*graphDataViewItem
}

type graphDataViewItem struct {
	Name    string
	Value   float64
	Percent float64
}

func GraphDataFromMap(in map[string]float64) *GraphDataSourceTable {
	ret := &GraphDataSourceTable{}

	var keys []string
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, v := range keys {
		ret.Items = append(ret.Items, &GraphDataSourceTableValue{
			Name:  v,
			Value: in[v],
		})
	}
	return ret
}

type GraphDataSourceTable struct {
	Items []*GraphDataSourceTableValue
}

type GraphDataSourceTableValue struct {
	Name  string
	Value float64
}

func (table *GraphDataSourceTable) Len() int {
	return len(table.Items)
}

func (table *GraphDataSourceTable) Name(i int) string {
	return table.Items[i].Name
}

func (table *GraphDataSourceTable) Value(i int) float64 {
	return table.Items[i].Value
}
