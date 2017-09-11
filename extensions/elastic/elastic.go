package elastic

import (
	"encoding/json"
	"fmt"
	"github.com/hypertornado/prago"
	administration "github.com/hypertornado/prago/extensions/admin"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
	"html/template"
)

func jsonize(in interface{}) template.HTML {
	result, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return template.HTML("")
	}
	return template.HTML(result)
}

var elasticMiddleware *ElasticMiddleware

type ElasticMiddleware struct {
	Admin  *administration.Admin
	client *elastic.Client
}

func (em *ElasticMiddleware) Init(app *prago.App) error {
	fmt.Println("init elastic middleware")
	elasticMiddleware = em

	client, err := elastic.NewClient()
	if err != nil {
		return err
	}

	em.client = client

	res, err := em.Admin.CreateResource(Elastic{})
	if err != nil {
		return err
	}
	res.HasModel = false
	res.Authenticate = administration.AuthenticateSysadmin

	return nil
}

type Elastic struct {
}

func (Elastic) InitResource(a *administration.Admin, resource *administration.Resource) error {

	resource.AddResourceAction(administration.ResourceAction{
		Url: "",
		Handler: func(a *administration.Admin, r *administration.Resource, request prago.Request) {

			urls := [][2]string{
				{"elastic/nodesinfo", "Nodes Info"},
				{"elastic/nodesstats", "Nodes Stats"},
			}

			indexes, err := elasticMiddleware.client.IndexNames()
			if err != nil {
				panic(err)
			}

			request.SetData("admin_title", "Elastic")
			request.SetData("admin_yield", "elastic_index")
			request.SetData("urls", urls)
			request.SetData("indexes", indexes)
			prago.Render(request, 200, "admin_layout")
		},
	})

	resource.AddResourceAction(administration.ResourceAction{
		Url: "nodesinfo",
		Handler: func(a *administration.Admin, r *administration.Resource, request prago.Request) {
			res, err := elasticMiddleware.client.NodesInfo().Do(context.Background())
			if err != nil {
				panic(err)
			}
			request.SetData("admin_title", "Nodes Info")
			request.SetData("data", jsonize(res))
			request.SetData("admin_yield", "elastic_pre")
			prago.Render(request, 200, "admin_layout")
		},
	})

	resource.AddResourceAction(administration.ResourceAction{
		Url: "nodesstats",
		Handler: func(a *administration.Admin, r *administration.Resource, request prago.Request) {
			res, err := elasticMiddleware.client.NodesStats().Do(context.Background())
			if err != nil {
				panic(err)
			}
			request.SetData("admin_title", "Nodes Stats")
			request.SetData("data", jsonize(res))
			request.SetData("admin_yield", "elastic_pre")
			prago.Render(request, 200, "admin_layout")
		},
	})

	resource.AddResourceAction(administration.ResourceAction{
		Url: "index/:name",
		Handler: func(a *administration.Admin, r *administration.Resource, request prago.Request) {

			name := request.Params().Get("name")

			/*fieldStats, err := elasticMiddleware.client.FieldStats(name).Do(context.Background())
			if err != nil {
				panic(err)
			}*/

			fieldStats, err := elasticMiddleware.client.FieldStats(name).Fields("Name").Do(context.Background())
			if err != nil {
				panic(err)
			}

			/*res, err := elasticMiddleware.client.NodesStats().Do(context.Background())
			if err != nil {
				panic(err)
			}*/
			request.SetData("admin_title", "Index "+name)
			request.SetData("index_name", name)

			request.SetData("fieldStats", fieldStats)

			request.SetData("admin_yield", "elastic_detail")
			prago.Render(request, 200, "admin_layout")
		},
	})

	return nil
}
