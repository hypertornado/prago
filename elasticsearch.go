package prago

import (
	"fmt"
	"sort"

	"github.com/hypertornado/prago/pragelastic"
)

func (app *App) initElasticsearch() {
	db := sysadminBoard.Dashboard(unlocalized("Elasticsearch"))
	db.AddTask(unlocalized("Reload elasticsearch client"), "sysadmin", func(ta *TaskActivity) error {
		return app.createNewElasticSearchClient()
	})

	ActionForm(app, "delete-elastic-indice", func(f *Form, r *Request) {
		stats, err := app.ElasticSearchClient().GetStats()
		if err != nil {
			panic(err)
		}

		var indiceNames []string
		for k := range stats.Indices {
			indiceNames = append(indiceNames, k)
		}
		sort.Strings(indiceNames)

		var doubled [][2]string = [][2]string{{"", ""}}

		for _, v := range indiceNames {
			doubled = append(doubled, [2]string{v, v})
		}

		f.AddSelect("indice", "Elastic indices", doubled)

		f.AddSubmit("Delete indice")
	}, func(vc FormValidation, request *Request) {
		id := request.Param("indice")
		if id == "" {
			vc.AddItemError("indice", "Select indice to delete")
		}
		if !vc.Valid() {
			return
		}

		err := app.ElasticSearchClient().DeleteIndex(id)
		if err != nil {
			vc.AddError(fmt.Sprintf("Index '%s' nelze smazat", id))
		} else {
			vc.AddError(fmt.Sprintf("Index '%s' úspěšně smazán", id))
		}
	}).Name(unlocalized("Smazat elasticsearch index")).Permission(sysadminPermission).Board(sysadminBoard)
}

var searchClient *pragelastic.Client

func (app *App) createNewElasticSearchClient() error {
	ret, err := pragelastic.New(app.codeName)
	if err != nil {
		return err
	}
	searchClient = ret
	return nil

}

func (app *App) ElasticSearchClient() *pragelastic.Client {
	if searchClient == nil {
		app.createNewElasticSearchClient()
	}
	return searchClient
}
