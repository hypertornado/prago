package prago

type Graph struct {
	name       string
	dataSource graphDataSource
}

func (t *Table) Graph() *Graph {
	graph := &Graph{}
	currentTable := t.currentTable()
	currentTable.Graphs = append(currentTable.Graphs, graph)
	return graph
}

func (g *Graph) Name(name string) *Graph {
	g.name = name
	return g
}

func (g *Graph) DataMap(m map[string]float64) *Graph {
	g.dataSource = graphDataFromMap(m)
	return g
}
