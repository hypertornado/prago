package prago

import (
	"context"
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

func (row *listStatsRow) GetTitle() string {
	var fields []string

	fields = append(fields, row.Name)
	fields = append(fields, row.Description.Count)
	if row.Description.Percent != "" {
		fields = append(fields, row.Description.Percent)
	}

	return strings.Join(fields, " — ")
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

func (resource *Resource) getListStats(ctx context.Context, userData UserData, params url.Values) *listStats {
	ret := &listStats{}

	columnsStr := params.Get("_columns")
	if columnsStr == "" {
		columnsStr = resource.defaultVisibleFieldsStr(userData)
	}
	columnsAr := strings.Split(columnsStr, ",")

	query := resource.addFilterParamsToQuery(resource.query(ctx), params, userData)
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
		field := resource.fieldMap[v]

		if field == nil {
			continue
		}

		if field.typ == reflect.TypeOf(time.Now()) {
			ret.Sections = append(ret.Sections, resource.getListStatsDateSections(ctx, field, userData, params, total, limit)...)
		}

		table := resource.getListStatsTable(ctx, field, userData, params, total, limit)

		if table == nil {
			continue
		}

		ret.Sections = append(ret.Sections, listStatsSection{
			Name:  field.name(userData.Locale()),
			Table: table,
		})
	}

	return ret
}

func (resource *Resource) getListStatsDateSections(ctx context.Context, field *Field, userData UserData, params url.Values, total, limit int64) (ret []listStatsSection) {
	ret = append(ret, resource.getListStatsDateSectionDay(ctx, field, userData, params, total, limit))
	ret = append(ret, resource.getListStatsDateSectionMonth(ctx, field, userData, params, total, limit))
	ret = append(ret, resource.getListStatsDateSectionYear(ctx, field, userData, params, total, limit))
	return
}

func (resource *Resource) getListStatsDateSectionDay(ctx context.Context, field *Field, userData UserData, params url.Values, total, limit int64) (ret listStatsSection) {
	query := resource.addFilterParamsToQuery(resource.query(ctx), params, userData)
	whereParams := query.values
	q := fmt.Sprintf("SELECT DAY(`%s`), MONTH(`%s`), YEAR(`%s`), COUNT(id) FROM %s %s GROUP BY DAY(`%s`), MONTH(`%s`), YEAR(`%s`) ORDER BY COUNT(id) DESC LIMIT %d;",
		field.id,
		field.id,
		field.id,
		resource.getID(),
		buildWhereString(query.conditions),
		field.id,
		field.id,
		field.id,
		limit,
	)
	rows, err := resource.app.db.QueryContext(ctx, q, whereParams...)
	if err != nil {
		fmt.Println(q)
		panic(err)
	}
	defer rows.Close()
	ret = listStatsSection{
		Name: fmt.Sprintf("%s – dny", field.name(userData.Locale())),
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

func (resource *Resource) getListStatsDateSectionMonth(ctx context.Context, field *Field, userData UserData, params url.Values, total, limit int64) (ret listStatsSection) {
	query := resource.addFilterParamsToQuery(resource.query(ctx), params, userData)
	whereParams := query.values
	q := fmt.Sprintf("SELECT MONTH(`%s`), YEAR(`%s`), COUNT(id) FROM %s %s GROUP BY MONTH(`%s`), YEAR(`%s`) ORDER BY COUNT(id) DESC LIMIT %d;",
		field.id,
		field.id,
		resource.getID(),
		buildWhereString(query.conditions),
		field.id,
		field.id,
		limit,
	)
	rows, err := resource.app.db.QueryContext(ctx, q, whereParams...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	ret = listStatsSection{
		Name: fmt.Sprintf("%s – měsíce", field.name(userData.Locale())),
	}
	var counted int64
	for rows.Next() {
		var month int64
		var year int64
		var count int64
		rows.Scan(&month, &year, &count)
		counted += count

		ret.Table = append(ret.Table, listStatsRow{
			Name:        fmt.Sprintf("%s %d", monthName(month, userData.Locale()), year),
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

func (resource *Resource) getListStatsDateSectionYear(ctx context.Context, field *Field, userData UserData, params url.Values, total, limit int64) (ret listStatsSection) {
	query := resource.addFilterParamsToQuery(resource.query(ctx), params, userData)
	whereParams := query.values
	q := fmt.Sprintf("SELECT YEAR(`%s`), COUNT(id) FROM %s %s GROUP BY YEAR(`%s`) ORDER BY COUNT(id) DESC LIMIT %d;",
		field.id,
		resource.getID(),
		buildWhereString(query.conditions),
		field.id,
		limit,
	)
	rows, err := resource.app.db.QueryContext(ctx, q, whereParams...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	ret = listStatsSection{
		Name: fmt.Sprintf("%s – roky", field.name(userData.Locale())),
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

func (resource *Resource) getListStatsTable(ctx context.Context, field *Field, userData UserData, params url.Values, total, limit int64) (table []listStatsRow) {
	query := resource.addFilterParamsToQuery(resource.query(ctx), params, userData)
	whereParams := query.values

	q := fmt.Sprintf("SELECT `%s`, COUNT(id) FROM %s %s GROUP BY `%s` ORDER BY COUNT(id) DESC LIMIT %d;", field.id, resource.getID(), buildWhereString(query.conditions), field.id, limit)

	rows, err := resource.app.db.QueryContext(ctx, q, whereParams...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var counted int64

	if field.typ.Kind() == reflect.String {
		for rows.Next() {
			var count int64
			var v string
			rows.Scan(&v, &count)
			counted += count

			var name = v
			if field.fieldType.isRelation() {
				name = humanizeMultiRelationsString(v)
			}

			table = append(table, listStatsRow{
				Name:        name,
				Description: statsCountDescription(count, total),
			})
		}
	}

	if field.typ.Kind() == reflect.Int64 {
		if field.fieldType.isRelation() {
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

				rd := field.relationPreview(userData, fmt.Sprintf("%d", v))
				if rd != nil {
					row.Name = rd[0].Name
					row.URL = rd[0].URL
					row.Image = rd[0].Image
				}

				table = append(table, row)
			}

		} else {
			table = resource.getListStatsTableInt(ctx, field, userData, params, total)
			counted = total
		}
	}

	if field.typ.Kind() == reflect.Float64 {
		table = resource.getListStatsTableInt(ctx, field, userData, params, total)
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

func (resource *Resource) getListStatsTableInt(ctx context.Context, field *Field, userData UserData, params url.Values, total int64) (table []listStatsRow) {
	if total <= 0 {
		return
	}

	query := resource.addFilterParamsToQuery(resource.query(ctx), params, userData)

	whereParams := query.values

	q := fmt.Sprintf("SELECT MAX(%s), MIN(%s), AVG(%s), SUM(%s) FROM %s %s;",
		field.id,
		field.id,
		field.id,
		field.id,
		resource.getID(),
		buildWhereString(query.conditions),
	)

	rows, err := resource.app.db.QueryContext(ctx, q, whereParams...)
	if err != nil {
		panic(err)
	}
	defer func() {
		rows.Close()
	}()

	var max float64
	var min float64
	var avg float64
	var sum float64

	for rows.Next() {
		rows.Scan(&max, &min, &avg, &sum)
	}
	//must(rows.Close())

	table = append(table, listStatsRow{
		Name: "minimum",
		Description: listStatsDescription{
			Count: humanizeFloat(min, userData.Locale()),
		},
	})

	table = append(table, listStatsRow{
		Name: "průměr",
		Description: listStatsDescription{
			Count: humanizeFloat(avg, userData.Locale()),
		},
	})

	medianItem := int64(math.Floor(float64(total) / 2))
	q = fmt.Sprintf("SELECT %s FROM %s %s LIMIT 1 OFFSET %d;",
		field.id,
		resource.getID(),
		buildWhereString(query.conditions),
		medianItem,
	)
	rows, err = resource.app.db.QueryContext(ctx, q, whereParams...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var median float64
	for rows.Next() {
		rows.Scan(&median)
	}
	table = append(table, listStatsRow{
		Name: "medián",
		Description: listStatsDescription{
			Count: humanizeFloat(median, userData.Locale()),
		},
	})
	table = append(table, listStatsRow{
		Name: "maximum",
		Description: listStatsDescription{
			Count: humanizeFloat(max, userData.Locale()),
		},
	})
	table = append(table, listStatsRow{
		Name: "součet",
		Description: listStatsDescription{
			Count: humanizeFloat(sum, userData.Locale()),
		},
	})

	return table
}

func getStatsLimitSelectData(locale string) (ret []listPaginationData) {
	var ints = []int64{10, 20, 100, 200, 500, 1000, 2000, 5000, 10000}

	for _, v := range ints {
		ret = append(ret, listPaginationData{
			Name:  messages.ItemsCount(v, locale),
			Value: v,
		})
	}

	ret[0].Selected = true
	return ret
}
