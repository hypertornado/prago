package administration

import (
	"fmt"
	"github.com/hypertornado/prago/administration/messages"
	"reflect"
	"time"
)

type StatsData struct {
	Fields []StatsField
}

type StatsField struct {
	Name     string
	Template string
	Data     interface{}
}

type StatsDataPie struct {
	ValueA int64
	LabelA string
	ValueB int64
	LabelB string
}

type StatsDataTimeline struct {
	Field    string
	Resource string
}

func (resource Resource) count() int64 {
	var item interface{}
	resource.newItem(&item)
	count, _ := resource.Admin.Query().Count(item)
	return count
}

func (resource Resource) getStats(user User) (ret StatsData) {
	ret.Fields = append(ret.Fields, StatsField{"𝚺", "admin_stats_text", fmt.Sprintf("%d", resource.count())})
	for _, v := range resource.fieldArrays {
		if defaultVisibilityFilter(resource, user, *v) {
			field := (*v).getStats(resource, user)
			if field != nil {
				ret.Fields = append(ret.Fields, *field)
			}
		}
	}
	return
}

func (f Field) getStats(resource Resource, user User) *StatsField {
	if f.Name == "ID" {
		return nil
	}

	switch f.Typ.Kind().String() {
	case "int", "int32", "int64":
		if f.Tags["prago-type"] == "relation" {
			return f.getStatsRelation(resource, user)
		}
		return f.getStatsInt(resource, user)
	case "string":
		return f.getStatsString(resource, user)
	case "bool":
		return f.getStatsBool(resource, user)
	}

	if f.Typ == reflect.TypeOf(time.Now()) {
		return &StatsField{
			Name:     f.HumanName(user.Locale),
			Template: "admin_stats_timeline",
			Data: StatsDataTimeline{
				Resource: resource.TableName,
				Field:    f.Name,
			},
		}
	}

	return nil
}

func (f Field) getStatsRelation(resource Resource, user User) *StatsField {
	var item interface{}
	resource.newItem(&item)
	zeroCount, _ := resource.Admin.Query().Where(fmt.Sprintf("`%s` is null or `%s` = 0", f.ColumnName, f.ColumnName)).Count(item)

	return f.getStatsFieldPie(
		messages.Messages.Get(user.Locale, "nonempty"),
		(resource.count() - zeroCount),
		messages.Messages.Get(user.Locale, "empty"),
		zeroCount,
		user,
	)
}

func (f Field) getStatsString(resource Resource, user User) *StatsField {
	var item interface{}
	resource.newItem(&item)
	zeroCount, _ := resource.Admin.Query().Where(fmt.Sprintf("`%s` is null or `%s` = \"\"", f.ColumnName, f.ColumnName)).Count(item)

	return f.getStatsFieldPie(
		messages.Messages.Get(user.Locale, "nonempty"),
		(resource.count() - zeroCount),
		messages.Messages.Get(user.Locale, "empty"),
		zeroCount,
		user,
	)
}

func (f Field) getStatsBool(resource Resource, user User) *StatsField {
	var item interface{}
	resource.newItem(&item)
	zeroCount, _ := resource.Admin.Query().Where(fmt.Sprintf("`%s` is null or `%s` = 0", f.ColumnName, f.ColumnName)).Count(item)

	return f.getStatsFieldPie(
		messages.Messages.Get(user.Locale, "yes"),
		(resource.count() - zeroCount),
		messages.Messages.Get(user.Locale, "no"),
		zeroCount,
		user,
	)
}

func (f Field) getStatsInt(resource Resource, user User) *StatsField {
	db := resource.Admin.GetDB()
	var max, min, sum, avg float64

	q := fmt.Sprintf("SELECT max(%s), min(%s), sum(%s), avg(%s) FROM %s", f.ColumnName, f.ColumnName, f.ColumnName, f.ColumnName, resource.TableName)
	must(db.QueryRow(q).Scan(&max, &min, &sum, &avg))
	return f.getStatsText(fmt.Sprintf("max: %v, min: %v, sum: %v, avg: %v", max, min, sum, avg), user.Locale)
}

func (f Field) getStatsText(text, language string) *StatsField {
	return &StatsField{
		Name:     f.HumanName(language),
		Template: "admin_stats_text",
		Data:     text,
	}
}

func (f Field) getStatsFieldPie(labelA string, valueA int64, labelB string, valueB int64, user User) *StatsField {

	return &StatsField{
		Name:     f.HumanName(user.Locale),
		Template: "admin_stats_pie",
		Data: StatsDataPie{
			ValueA: valueA,
			LabelA: labelA,
			ValueB: valueB,
			LabelB: labelB,
		},
	}
}
