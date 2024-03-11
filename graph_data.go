package prago

import (
	"sort"
)

type graphDataSource interface {
	Len() int
	Name(int) string
	Value(int) float64
}

func (graph *Graph) View() *graphDataView {

	ret := &graphDataView{
		Name: graph.name,
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

func graphDataFromMap(in map[string]float64) *graphDataSourceTable {
	ret := &graphDataSourceTable{}

	var keys []string
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, v := range keys {
		ret.Items = append(ret.Items, &graphDataSourceTableValue{
			Name:  v,
			Value: in[v],
		})
	}
	return ret
}

type graphDataSourceTable struct {
	Items []*graphDataSourceTableValue
}

type graphDataSourceTableValue struct {
	Name  string
	Value float64
}

func (table *graphDataSourceTable) Len() int {
	return len(table.Items)
}

func (table *graphDataSourceTable) Name(i int) string {
	return table.Items[i].Name
}

func (table *graphDataSourceTable) Value(i int) float64 {
	return table.Items[i].Value
}
