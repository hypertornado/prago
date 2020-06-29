package administration

import (
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strings"

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

	/*return fmt.Sprintf("%s (%s)",
		utils.HumanizeNumber(count),
		statsCountPercent(count, total),
	)*/
}

func getListStats(resource *Resource, admin *Administration, user User, params url.Values) *ListStats {
	ret := &ListStats{}

	columnsStr := params.Get("_columns")
	if columnsStr == "" {
		columnsStr = resource.defaultVisibleFieldsStr()
	}
	columnsAr := strings.Split(columnsStr, ",")

	query := resource.addFilterParamsToQuery(admin.Query(), params)

	var item interface{}
	resource.newItem(&item)
	total, err := query.Count(item)
	if err != nil {
		panic(err)
	}

	for _, v := range columnsAr {

		field := resource.fieldMap[v]

		//query := admin.Query()

		/*for _, field := range resource.fieldArrays {
			val := params.Get(field.ColumnName)
			if val != "" {
				query = query.WhereIs(field.ColumnName, val)
			}
		}*/

		query := resource.addFilterParamsToQuery(admin.Query(), params)

		whereParams := query.query.whereParams

		q := fmt.Sprintf("SELECT %s, COUNT(id) FROM %s %s GROUP BY %s ORDER BY COUNT(id) DESC LIMIT 10;", field.ColumnName, resource.TableName, buildWhereString(query.query.whereString), field.ColumnName)

		rows, err := admin.db.Query(q, whereParams...)
		if err != nil {
			panic(err)
		}

		var counted int64

		var table []ListStatsRow

		if field.Typ.Kind() == reflect.String {
			for rows.Next() {
				var count int64
				var v string
				rows.Scan(&v, &count)
				counted += count

				table = append(table, ListStatsRow{
					Name: v,
					//Percent:     statsCountPercent(count, total),
					Description: statsCountDescription(count, total),
				})
			}
		}

		if field.Typ.Kind() == reflect.Int64 {
			for rows.Next() {
				var count int64
				var v int64
				rows.Scan(&v, &count)
				counted += count

				table = append(table, ListStatsRow{
					Name: fmt.Sprintf("%d", v),
					//Percent:     statsCountPercent(count, total),
					Description: statsCountDescription(count, total),
				})
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
				Name: "ano",
				//Percent:     statsCountPercent(countTrue, total),
				Description: statsCountDescription(countTrue, total),
			})
			table = append(table, ListStatsRow{
				Name: "ne",
				//Percent:     statsCountPercent(countFalse, total),
				Description: statsCountDescription(countFalse, total),
			})
		}

		if counted < total {
			table = append(table, ListStatsRow{
				Name: "ostatní",
				//Percent:     statsCountPercent(total-counted, total),
				Description: statsCountDescription(total-counted, total),
			})
		}

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
