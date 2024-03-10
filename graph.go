package prago

type graph struct {
	Name       string
	dataSource graphDataSource
	//Values     []*graphValue
}

/*type graphValue struct {
	Name  string
	Value float64
	graph *graph
}*/

func (t *Table) Graph(name string, graphDataSource graphDataSource) {
	graph := &graph{
		Name:       name,
		dataSource: graphDataSource,
	}

	/*for _, v := range graph.Values {
		v.graph = graph
	}*/

	currentTable := t.currentTable()
	currentTable.Graphs = append(currentTable.Graphs, graph)
}
