package extensions

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/hypertornado/prago/utils"
	"reflect"
	"strings"
)

type AdminResource struct {
	ID    string
	Name  string
	Item  interface{}
	Items []*AdminResourceItem
	admin *Admin
}

func NewResource(item interface{}) (*AdminResource, error) {
	name := reflect.TypeOf(item).Name()

	ret := &AdminResource{
		Item: item,
		Name: name,
		ID:   utils.PrettyUrl(name),
	}
	return ret, nil
}

type mysqlColumn struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   sql.NullString
}

func (ar *AdminResource) db() *sql.DB {
	return ar.admin.db
}

func (ar *AdminResource) tableName() string {
	return ar.ID
}

func (ar *AdminResource) dropTable() error {
	_, err := ar.db().Exec(fmt.Sprintf("drop table `%s`;", ar.tableName()))
	return err
}

func (ar *AdminResource) getTableDescription() (map[string]*mysqlColumn, error) {
	fmt.Println("Describe " + ar.tableName())

	columns := map[string]*mysqlColumn{}

	rows, err := ar.db().Query(fmt.Sprintf("describe `%s`;", ar.tableName()))
	if err != nil {
		return columns, err
	}
	defer rows.Close()

	for rows.Next() {
		column := &mysqlColumn{}
		rows.Scan(
			&column.Field,
			&column.Type,
			&column.Null,
			&column.Key,
			&column.Default,
			&column.Extra,
		)
		columns[column.Field] = column
	}

	return columns, nil
}

func (ar *AdminResource) getStructDescription() (map[string]*mysqlColumn, error) {
	columns := map[string]*mysqlColumn{}

	typ := reflect.TypeOf(ar.Item)
	for i := 0; i < typ.NumField(); i++ {
		use := true
		field := typ.Field(i)
		column := &mysqlColumn{
			Field: utils.PrettyUrl(field.Name),
		}

		switch field.Type.Kind().String() {
		case "int64":
			column.Type = "bigint(20)"
		case "string":
			column.Type = "varchar(255)"
		default:
			fmt.Println("Cant use field", field.Name)
			use = false
		}
		if use {
			columns[column.Field] = column
		}
	}
	return columns, nil
}

type listResult struct {
	Url  string
	Name string
}

func (ar *AdminResource) List() ([]listResult, error) {
	ret := []listResult{}
	q := fmt.Sprintf("SELECT id, name FROM `%s`;", ar.tableName())

	rows, err := ar.db().Query(q)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		rows.Scan(&id, &name)
		item := &listResult{
			Url:  fmt.Sprintf("%s/%s/%d", ar.admin.Prefix, ar.ID, id),
			Name: name,
		}
		ret = append(ret, *item)
	}

	return ret, nil
}

func (ar *AdminResource) CreateItem(item interface{}) error {
	typ := reflect.TypeOf(item)
	value := reflect.ValueOf(item)

	id := utils.PrettyUrl(typ.Name())

	if id != ar.tableName() {
		return errors.New("Wrong class of item")
	}

	names := []string{}
	questionMarks := []string{}
	values := []interface{}{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Name == "ID" {
			continue
		}

		val := value.FieldByName(field.Name)

		switch field.Type.Kind() {
		case reflect.String:
			values = append(values, val.String())
		default:
			continue
		}

		names = append(names, "`"+utils.PrettyUrl(field.Name)+"`")
		questionMarks = append(questionMarks, "?")

	}

	q := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);", ar.tableName(), strings.Join(names, ", "), strings.Join(questionMarks, ", "))
	_, err := ar.db().Exec(q, values...)

	return err
}

func (ar *AdminResource) createTable(description map[string]*mysqlColumn) error {
	items := []string{}

	for _, v := range description {
		additional := ""
		if v.Field == "id" {
			additional = "NOT NULL AUTO_INCREMENT PRIMARY KEY"
		}
		item := fmt.Sprintf("%s %s %s", v.Field, v.Type, additional)
		items = append(items, item)
	}

	q := fmt.Sprintf("CREATE TABLE %s (%s);", ar.tableName(), strings.Join(items, ", "))
	_, err := ar.db().Exec(q)
	return err
}

func (ar *AdminResource) Migrate() error {

	var err error

	fmt.Println("Migrating ", ar.Name, ar.ID)

	err = ar.dropTable()
	if err != nil {
		fmt.Println(err)
	}

	_, err = ar.getTableDescription()

	structDescription, err := ar.getStructDescription()
	if err != nil {
		return err
	}
	return ar.createTable(structDescription)
}
