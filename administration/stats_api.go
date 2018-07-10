package administration

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/hypertornado/prago"
	"io/ioutil"
	"time"
)

type StatsAPIData struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
}

type StatsAPIResult struct {
	Labels []string `json:"labels"`
	Values []int64  `json:"values"`
}

//select month(createdat) as m, year(createdat) as y, count(*) from packageorder where createdat >= '2018-02-01' AND createdat <= '2018-12-01' group by m, y order by y desc, m desc;

func bindStatsAPI(admin *Administration) {
	admin.AdminController.Post(admin.GetURL("_api/stats"), func(request prago.Request) {

		user := GetUser(request)

		b, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic(err)
		}

		var data StatsAPIData
		err = json.Unmarshal(b, &data)
		if err != nil {
			panic(err)
		}

		resource := admin.getResourceByName(data.Resource)
		if resource == nil {
			panic("can't find resource")
		}

		if !admin.Authorize(user, resource.CanView) {
			render403(request)
			return
		}

		field := resource.fieldMap[data.Field]

		end := time.Now()
		start := end.AddDate(-1, 0, -1)

		q := fmt.Sprintf(
			"select month(%s) as m, year(%s) as y, count(*) from `%s` where createdat >= '%s' AND createdat <= '%s' group by m, y order by y desc, m desc;",
			field.ColumnName,
			field.ColumnName,
			resource.TableName,
			start.Format("2006-01-02"),
			end.Format("2006-01-02"),
		)

		var rows *sql.Rows
		rows, err = admin.db.Query(q)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		resultsMap := map[[2]int64]int64{}

		for rows.Next() {
			var month, year, count int64
			err = rows.Scan(&month, &year, &count)
			if err != nil {
				panic(err)
			}
			resultsMap[[2]int64{month, year}] = count
		}

		var ret StatsAPIResult

		t := start
		for t.Before(end) {
			ret.Labels = append(ret.Labels, t.Format("2006-01"))
			ret.Values = append(ret.Values, resultsMap[[2]int64{int64(t.Month()), int64(t.Year())}])
			t = t.AddDate(0, 1, 0)
		}

		request.RenderJSON(ret)
	})
}
