package prago

/*
import (
	"bytes"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

//https://github.com/goccy/go-graphviz/pull/77

func (app *App) initResourceConnections() {

	app.API("resource-connections").Method("GET").Permission("sysadmin").Handler(func(request *Request) {
		request.Response().Header().Add("Content-Type", "image/svg+xml")

		g := graphviz.New()
		graph, err := g.Graph()
		must(err)
		defer func() {
			if err := graph.Close(); err != nil {
				log.Fatal(err)
			}
			g.Close()
		}()

		nodeMap := make(map[string]*cgraph.Node)

		for _, resource := range app.resources {
			n, err := graph.CreateNode(resource.pluralName("en"))
			n.SetShape(cgraph.RectShape)
			must(err)
			nodeMap[resource.id] = n
		}

		for _, resource := range app.resources {
			for _, field := range resource.fields {
				if field.relatedResource != nil {
					edge, err := graph.CreateEdge(field.name("en"), nodeMap[resource.id], nodeMap[field.relatedResource.id])
					edge.SetColor("#aaaaaa")
					edge.SetLabelFontColor("#aaaaaa")
					edge.SetFontColor("#aaaaaa")
					must(err)
					edge.SetLabel(field.name("en"))
				}
			}

		}

		var buf bytes.Buffer
		if err := g.Render(graph, graphviz.SVG, &buf); err != nil {
			log.Fatal(err)
		}
		request.Response().Write(buf.Bytes())

	})

	app.Action("resource-connections").Board(sysadminBoard).Permission("sysadmin").Name(unlocalized("Resource connections")).View("admin_resource_connections", func(r *Request) any {
		return nil
	})

}
*/
