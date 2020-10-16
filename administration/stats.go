package administration

import (
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/utils"
)

type ListStats struct {
	Sections []ListStatsSection
}

type ListStatsSection struct {
	Name  string
	Table []ListStatsRow
}

type ListStatsRow struct {
	Name        string
	Image       string
	URL         string
	Description ListStatsDescription
}

type ListStatsDescription struct {
	Count      string
	PercentCSS template.HTML
	Percent    string
}

func statsCountPercent(count, total int64) template.HTML {
	if total == 0 {
		return ""
	}
	return template.HTML(fmt.Sprintf("%.2f%%", (100*float64(count))/float64(total)))
}

func statsCountDescription(count, total int64) ListStatsDescription {
	percentStr := statsCountPercent(count, total)
	return ListStatsDescription{
		Count:      utils.HumanizeNumber(count),
		PercentCSS: percentStr,
		Percent:    string(percentStr),
	}
}

func getListStats(resource *Resource, user User, params url.Values) *ListStats {
	ret := &ListStats{}

	columnsStr := params.Get("_columns")
	if columnsStr == "" {
		columnsStr = resource.defaultVisibleFieldsStr()
	}
	columnsAr := strings.Split(columnsStr, ",")

	query := resource.addFilterParamsToQuery(resource.Admin.Query(), params)

	var item interface{}
	resource.newItem(&item)
	total, err := query.Count(item)
	if err != nil {
		panic(err)
	}

	var limit int64 = 10
	statsLimit, err := strconv.Atoi(params.Get("_statslimit"))
	if err == nil && statsLimit > 0 && statsLimit <= 10000 {
		limit = int64(statsLimit)
	}

	for _, v := range columnsAr {

		field := resource.fieldMap[v]

		table := resource.getListStatsTable(field, user, params, total, limit)

		if table == nil {
			continue
		}

		ret.Sections = append(ret.Sections, ListStatsSection{
			Name:  field.HumanName(user.Locale),
			Table: table,
		})
	}

	return ret
}

func (resource *Resource) getListStatsTable(field *Field, user User, params url.Values, total, limit int64) (table []ListStatsRow) {

	query := resource.addFilterParamsToQuery(resource.Admin.Query(), params)

	whereParams := query.query.whereParams

	q := fmt.Sprintf("SELECT %s, COUNT(id) FROM %s %s GROUP BY %s ORDER BY COUNT(id) DESC LIMIT %d;", field.ColumnName, resource.TableName, buildWhereString(query.query.whereString), field.ColumnName, limit)

	rows, err := resource.Admin.db.Query(q, whereParams...)
	if err != nil {
		panic(err)
	}

	var counted int64

	if field.Typ.Kind() == reflect.String {
		for rows.Next() {
			var count int64
			var v string
			rows.Scan(&v, &count)
			counted += count

			table = append(table, ListStatsRow{
				Name:        v,
				Description: statsCountDescription(count, total),
			})
		}
	}

	if field.Typ.Kind() == reflect.Int64 {

		if field.fieldType.IsRelation() {
			for rows.Next() {
				var count int64
				var v int64
				rows.Scan(&v, &count)
				counted += count

				row := ListStatsRow{
					Name:        fmt.Sprintf("#%d", v),
					Description: statsCountDescription(count, total),
				}

				if v == 0 {
					row.Name = "–"
				}

				rd, err := getRelationData(*resource, user, *field, v)
				if err == nil {
					row.Name = rd.Name
					row.URL = rd.URL
					row.Image = rd.Image
				}

				table = append(table, row)
			}

		} else {

			table = resource.getListStatsTableInt(field, user, params, total)

			counted = total

			/*for rows.Next() {
				q := fmt.Sprintf("SELECT %s, COUNT(id) FROM %s %s GROUP BY %s ORDER BY COUNT(id) DESC LIMIT %d;", field.ColumnName, resource.TableName, buildWhereString(query.query.whereString), field.ColumnName, limit)

				rows, err := resource.Admin.db.Query(q, whereParams...)
				if err != nil {
					panic(err)
				}

				var count int64
				var v int64
				rows.Scan(&v, &count)
				counted += count

				table = append(table, ListStatsRow{
					Name:        fmt.Sprintf("%d", v),
					Description: statsCountDescription(count, total),
				})
			}*/
		}
	}

	if field.Typ.Kind() == reflect.Bool {
		var countTrue, countFalse int64

		for rows.Next() {
			var count int64
			var v bool
			rows.Scan(&v, &count)
			counted += count

			if v {
				countTrue = count
			} else {
				countFalse = count
			}

		}

		table = append(table, ListStatsRow{
			Name:        "ano",
			Description: statsCountDescription(countTrue, total),
		})
		table = append(table, ListStatsRow{
			Name:        "ne",
			Description: statsCountDescription(countFalse, total),
		})
	}

	if counted < total && len(table) > 0 {
		table = append(table, ListStatsRow{
			Name:        "ostatní",
			Description: statsCountDescription(total-counted, total),
		})
	}
	return
}

func (resource *Resource) getListStatsTableInt(field *Field, user User, params url.Values, total int64) (table []ListStatsRow) {
	if total <= 0 {
		return
	}

	query := resource.addFilterParamsToQuery(resource.Admin.Query(), params)

	whereParams := query.query.whereParams

	q := fmt.Sprintf("SELECT MAX(%s), MIN(%s), AVG(%s) FROM %s %s;",
		field.ColumnName,
		field.ColumnName,
		field.ColumnName,
		resource.TableName,
		buildWhereString(query.query.whereString),
	)

	rows, err := resource.Admin.db.Query(q, whereParams...)
	if err != nil {
		panic(err)
	}

	var max int64
	var min int64
	var avg float64

	for rows.Next() {
		rows.Scan(&max, &min, &avg)
		//counted += count
	}

	table = append(table, ListStatsRow{
		Name: "minimum",
		Description: ListStatsDescription{
			Count: utils.HumanizeNumber(min),
		},
	})

	table = append(table, ListStatsRow{
		Name: "průměr",
		Description: ListStatsDescription{
			Count: fmt.Sprintf("%.2f", avg),
		},
	})

	table = append(table, ListStatsRow{
		Name: "maximum",
		Description: ListStatsDescription{
			Count: utils.HumanizeNumber(max),
		},
	})

	return table
}

func getStatsLimitSelectData(locale string) (ret []ListPaginationData) {
	var ints []int64 = []int64{5, 10, 20, 100, 200, 500, 1000, 2000, 5000, 10000}

	for _, v := range ints {
		ret = append(ret, ListPaginationData{
			Name:  messages.Messages.ItemsCount(v, locale),
			Value: v,
		})
	}

	ret[0].Selected = true
	return ret
}
