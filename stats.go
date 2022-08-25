package prago

import (
	"fmt"
	"html/template"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type listStats struct {
	Sections []listStatsSection
}

type listStatsSection struct {
	Name  string
	Table []listStatsRow
}

type listStatsRow struct {
	Name        string
	Image       string
	URL         string
	Description listStatsDescription
}

type listStatsDescription struct {
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

func statsCountDescription(count, total int64) listStatsDescription {
	percentStr := statsCountPercent(count, total)
	return listStatsDescription{
		Count:      humanizeNumber(count),
		PercentCSS: percentStr,
		Percent:    string(percentStr),
	}
}

func (resourceData *resourceData) getListStats(user *user, params url.Values) *listStats {
	ret := &listStats{}

	columnsStr := params.Get("_columns")
	if columnsStr == "" {
		columnsStr = resourceData.defaultVisibleFieldsStr(user)
	}
	columnsAr := strings.Split(columnsStr, ",")

	query := resourceData.addFilterParamsToQuery(resourceData.query(), params)
	total, err := query.count()
	if err != nil {
		panic(err)
	}

	var limit int64 = 10
	statsLimit, err := strconv.Atoi(params.Get("_statslimit"))
	if err == nil && statsLimit > 0 && statsLimit <= 10000 {
		limit = int64(statsLimit)
	}

	for _, v := range columnsAr {
		field := resourceData.fieldMap[v]

		if field.typ == reflect.TypeOf(time.Now()) {
			ret.Sections = append(ret.Sections, resourceData.getListStatsDateSections(field, user, params, total, limit)...)
		}

		table := resourceData.getListStatsTable(field, user, params, total, limit)

		if table == nil {
			continue
		}

		ret.Sections = append(ret.Sections, listStatsSection{
			Name:  field.name(user.Locale),
			Table: table,
		})
	}

	return ret
}

func (resourceData *resourceData) getListStatsDateSections(field *Field, user *user, params url.Values, total, limit int64) (ret []listStatsSection) {
	ret = append(ret, resourceData.getListStatsDateSectionDay(field, user, params, total, limit))
	ret = append(ret, resourceData.getListStatsDateSectionMonth(field, user, params, total, limit))
	ret = append(ret, resourceData.getListStatsDateSectionYear(field, user, params, total, limit))
	return
}

func (resourceData *resourceData) getListStatsDateSectionDay(field *Field, user *user, params url.Values, total, limit int64) (ret listStatsSection) {
	query := resourceData.addFilterParamsToQuery(resourceData.query(), params)
	whereParams := query.values
	q := fmt.Sprintf("SELECT DAY(%s), MONTH(%s), YEAR(%s), COUNT(id) FROM %s %s GROUP BY DAY(%s), MONTH(%s), YEAR(%s) ORDER BY COUNT(id) DESC LIMIT %d;",
		field.id,
		field.id,
		field.id,
		resourceData.getID(),
		buildWhereString(query.conditions),
		field.id,
		field.id,
		field.id,
		limit,
	)
	rows, err := resourceData.app.db.Query(q, whereParams...)
	if err != nil {
		panic(err)
	}
	ret = listStatsSection{
		Name: fmt.Sprintf("%s – dny", field.name(user.Locale)),
	}
	var counted int64
	for rows.Next() {
		var day int64
		var month int64
		var year int64
		var count int64
		rows.Scan(&day, &month, &year, &count)
		counted += count

		ret.Table = append(ret.Table, listStatsRow{
			Name:        fmt.Sprintf("%d. %d. %d", day, month, year),
			Description: statsCountDescription(count, total),
		})
	}
	if counted < total {
		ret.Table = append(ret.Table, listStatsRow{
			Name:        "ostatní",
			Description: statsCountDescription(total-counted, total),
		})
	}
	return
}

func (resourceData *resourceData) getListStatsDateSectionMonth(field *Field, user *user, params url.Values, total, limit int64) (ret listStatsSection) {
	query := resourceData.addFilterParamsToQuery(resourceData.query(), params)
	whereParams := query.values
	q := fmt.Sprintf("SELECT MONTH(%s), YEAR(%s), COUNT(id) FROM %s %s GROUP BY MONTH(%s), YEAR(%s) ORDER BY COUNT(id) DESC LIMIT %d;",
		field.id,
		field.id,
		resourceData.getID(),
		buildWhereString(query.conditions),
		field.id,
		field.id,
		limit,
	)
	rows, err := resourceData.app.db.Query(q, whereParams...)
	if err != nil {
		panic(err)
	}
	ret = listStatsSection{
		Name: fmt.Sprintf("%s – měsíce", field.name(user.Locale)),
	}
	var counted int64
	for rows.Next() {
		var month int64
		var year int64
		var count int64
		rows.Scan(&month, &year, &count)
		counted += count

		ret.Table = append(ret.Table, listStatsRow{
			Name:        fmt.Sprintf("%s %d", monthName(month, user.Locale), year),
			Description: statsCountDescription(count, total),
		})
	}
	if counted < total {
		ret.Table = append(ret.Table, listStatsRow{
			Name:        "ostatní",
			Description: statsCountDescription(total-counted, total),
		})
	}
	return
}

func (resourceData *resourceData) getListStatsDateSectionYear(field *Field, user *user, params url.Values, total, limit int64) (ret listStatsSection) {
	query := resourceData.addFilterParamsToQuery(resourceData.query(), params)
	whereParams := query.values
	q := fmt.Sprintf("SELECT YEAR(%s), COUNT(id) FROM %s %s GROUP BY YEAR(%s) ORDER BY COUNT(id) DESC LIMIT %d;",
		field.id,
		resourceData.getID(),
		buildWhereString(query.conditions),
		field.id,
		limit,
	)
	rows, err := resourceData.app.db.Query(q, whereParams...)
	if err != nil {
		panic(err)
	}
	ret = listStatsSection{
		Name: fmt.Sprintf("%s – roky", field.name(user.Locale)),
	}
	var counted int64
	for rows.Next() {
		var year int64
		var count int64
		rows.Scan(&year, &count)
		counted += count

		ret.Table = append(ret.Table, listStatsRow{
			Name:        fmt.Sprintf("%d", year),
			Description: statsCountDescription(count, total),
		})
	}
	if counted < total {
		ret.Table = append(ret.Table, listStatsRow{
			Name:        "ostatní",
			Description: statsCountDescription(total-counted, total),
		})
	}
	return
}

func (resourceData *resourceData) getListStatsTable(field *Field, user *user, params url.Values, total, limit int64) (table []listStatsRow) {
	query := resourceData.addFilterParamsToQuery(resourceData.query(), params)
	whereParams := query.values

	q := fmt.Sprintf("SELECT %s, COUNT(id) FROM %s %s GROUP BY %s ORDER BY COUNT(id) DESC LIMIT %d;", field.id, resourceData.getID(), buildWhereString(query.conditions), field.id, limit)

	rows, err := resourceData.app.db.Query(q, whereParams...)
	if err != nil {
		panic(err)
	}

	var counted int64

	if field.typ.Kind() == reflect.String {
		for rows.Next() {
			var count int64
			var v string
			rows.Scan(&v, &count)
			counted += count

			table = append(table, listStatsRow{
				Name:        v,
				Description: statsCountDescription(count, total),
			})
		}
	}

	if field.typ.Kind() == reflect.Int64 {
		if field.fieldType.IsRelation() {
			for rows.Next() {
				var count int64
				var v int64
				rows.Scan(&v, &count)
				counted += count

				row := listStatsRow{
					Name:        fmt.Sprintf("#%d", v),
					Description: statsCountDescription(count, total),
				}

				if v == 0 {
					row.Name = "–"
				}

				rd, err := getPreviewData(user, field, v)
				if err == nil {
					row.Name = rd.Name
					row.URL = rd.URL
					row.Image = rd.Image
				}

				table = append(table, row)
			}

		} else {
			table = resourceData.getListStatsTableInt(field, user, params, total)
			counted = total
		}
	}

	if field.typ.Kind() == reflect.Float64 {
		table = resourceData.getListStatsTableInt(field, user, params, total)
		counted = total
	}

	if field.typ.Kind() == reflect.Bool {
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

		table = append(table, listStatsRow{
			Name:        "ano",
			Description: statsCountDescription(countTrue, total),
		})
		table = append(table, listStatsRow{
			Name:        "ne",
			Description: statsCountDescription(countFalse, total),
		})
	}

	if counted < total && len(table) > 0 {
		table = append(table, listStatsRow{
			Name:        "ostatní",
			Description: statsCountDescription(total-counted, total),
		})
	}
	return
}

func (resourceData *resourceData) getListStatsTableInt(field *Field, user *user, params url.Values, total int64) (table []listStatsRow) {
	if total <= 0 {
		return
	}

	query := resourceData.addFilterParamsToQuery(resourceData.query(), params)

	whereParams := query.values

	q := fmt.Sprintf("SELECT MAX(%s), MIN(%s), AVG(%s), SUM(%s) FROM %s %s;",
		field.id,
		field.id,
		field.id,
		field.id,
		resourceData.getID(),
		buildWhereString(query.conditions),
	)

	rows, err := resourceData.app.db.Query(q, whereParams...)
	if err != nil {
		panic(err)
	}

	var max float64
	var min float64
	var avg float64
	var sum float64

	for rows.Next() {
		rows.Scan(&max, &min, &avg, &sum)
	}

	table = append(table, listStatsRow{
		Name: "minimum",
		Description: listStatsDescription{
			Count: humanizeFloat(min, user.Locale),
		},
	})

	table = append(table, listStatsRow{
		Name: "průměr",
		Description: listStatsDescription{
			Count: humanizeFloat(avg, user.Locale),
		},
	})

	medianItem := int64(math.Floor(float64(total) / 2))
	q = fmt.Sprintf("SELECT %s FROM %s %s LIMIT 1 OFFSET %d;",
		field.id,
		resourceData.getID(),
		buildWhereString(query.conditions),
		medianItem,
	)
	rows, err = resourceData.app.db.Query(q, whereParams...)
	if err != nil {
		panic(err)
	}
	var median float64
	for rows.Next() {
		rows.Scan(&median)
	}
	table = append(table, listStatsRow{
		Name: "medián",
		Description: listStatsDescription{
			Count: humanizeFloat(median, user.Locale),
		},
	})
	table = append(table, listStatsRow{
		Name: "maximum",
		Description: listStatsDescription{
			Count: humanizeFloat(max, user.Locale),
		},
	})
	table = append(table, listStatsRow{
		Name: "součet",
		Description: listStatsDescription{
			Count: humanizeFloat(sum, user.Locale),
		},
	})

	return table
}

func getStatsLimitSelectData(locale string) (ret []listPaginationData) {
	var ints = []int64{5, 10, 20, 100, 200, 500, 1000, 2000, 5000, 10000}

	for _, v := range ints {
		ret = append(ret, listPaginationData{
			Name:  messages.ItemsCount(v, locale),
			Value: v,
		})
	}

	ret[0].Selected = true
	return ret
}
