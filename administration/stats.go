package administration

import (
	"fmt"
)

type StatsData struct {
	Table [][2]string
}

func (resource Resource) count() int64 {
	var item interface{}
	resource.newItem(&item)
	count, _ := resource.Admin.Query().Count(item)
	return count
}

func (resource Resource) getStats(user User) StatsData {

	table := [][2]string{}

	table = append(table, [2]string{"𝚺", fmt.Sprintf("%d", resource.count())})

	for _, v := range resource.fieldArrays {
		if defaultVisibilityFilter(resource, user, *v) {
			table = append(table, (*v).getStats(resource, user)...)
		}
	}

	return StatsData{
		Table: table,
	}
}

func (f Field) getStats(resource Resource, user User) (ret [][2]string) {
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
	return nil
}

func (f Field) getStatsRelation(resource Resource, user User) (ret [][2]string) {
	var item interface{}
	resource.newItem(&item)
	zeroCount, _ := resource.Admin.Query().Where(fmt.Sprintf("`%s` is null or `%s` = 0", f.ColumnName, f.ColumnName)).Count(item)
	return [][2]string{{f.HumanName(user.Locale), fmt.Sprintf("empty: %v, nonempty: %v", zeroCount, (resource.count() - zeroCount))}}
}

func (f Field) getStatsString(resource Resource, user User) (ret [][2]string) {
	var item interface{}
	resource.newItem(&item)
	zeroCount, _ := resource.Admin.Query().Where(fmt.Sprintf("`%s` is null or `%s` = \"\"", f.ColumnName, f.ColumnName)).Count(item)
	return [][2]string{{f.HumanName(user.Locale), fmt.Sprintf("empty: %v, nonempty: %v", zeroCount, (resource.count() - zeroCount))}}
}

func (f Field) getStatsBool(resource Resource, user User) (ret [][2]string) {
	var item interface{}
	resource.newItem(&item)
	zeroCount, _ := resource.Admin.Query().Where(fmt.Sprintf("`%s` is null or `%s` = 0", f.ColumnName, f.ColumnName)).Count(item)
	return [][2]string{{f.HumanName(user.Locale), fmt.Sprintf("false: %v, true: %v", zeroCount, (resource.count() - zeroCount))}}
}

func (f Field) getStatsInt(resource Resource, user User) (ret [][2]string) {
	db := resource.Admin.GetDB()
	var max, min, sum, avg float64

	q := fmt.Sprintf("SELECT max(%s), min(%s), sum(%s), avg(%s) FROM %s", f.ColumnName, f.ColumnName, f.ColumnName, f.ColumnName, resource.TableName)
	must(db.QueryRow(q).Scan(&max, &min, &sum, &avg))

	return [][2]string{
		{
			f.Name,
			fmt.Sprintf("max: %v, min: %v, sum: %v, avg: %v", max, min, sum, avg),
		},
	}
}
