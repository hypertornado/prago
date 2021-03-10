package prago

import (
	"database/sql"
	"fmt"
	"strings"
)

func (app *App) initMigrationCommand() {
	app.AddCommand("admin", "migrate").Description("migrate database").
		Callback(func() {
			app.Log().Println("Migrating database")
			err := app.migrate(true)
			if err == nil {
				app.Log().Println("Migration done")
			} else {
				app.Log().Fatal(err)
			}
		})
}

func (app *App) migrate(verbose bool) error {
	tables, err := listTables(app.db)
	if err != nil {
		return err
	}
	for _, resource := range app.resources {
		tables[resource.TableName] = false
		err := resource.migrate(verbose)
		if err != nil {
			return err
		}
	}

	if verbose {
		unusedTables := []string{}
		for k, v := range tables {
			if v == true {
				unusedTables = append(unusedTables, k)
			}
		}
		if len(unusedTables) > 0 {
			fmt.Printf("Unused tables: %s\n", strings.Join(unusedTables, ", "))
		}
	}

	return nil
}

func (resource *Resource) unsafeDropTable() error {
	_, err := resource.App.db.Exec(fmt.Sprintf("drop table `%s`;", resource.TableName))
	return err
}

func (resource *Resource) migrate(verbose bool) error {
	_, err := getTableDescription(resource.App.db, resource.TableName)
	if err == nil {
		return migrateTable(resource.App.db, resource.TableName, *resource, verbose)
	}
	return createTable(resource.App.db, resource.TableName, *resource, verbose)
}

func listTables(db dbIface) (ret map[string]bool, err error) {
	ret = make(map[string]bool)
	var rows *sql.Rows
	rows, err = db.Query("show tables;")
	defer rows.Close()
	if err != nil {
		return ret, err
	}

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return
		}
		ret[name] = true
	}
	return
}

func createTable(db dbIface, tableName string, resource Resource, verbose bool) (err error) {
	if verbose {
		fmt.Printf("Creating table '%s'\n", tableName)
	}
	items := []string{}
	for _, v := range resource.fieldArrays {
		items = append(items, v.fieldDescriptionMysql(resource.fieldTypes))
	}
	q := fmt.Sprintf("CREATE TABLE %s (%s);", tableName, strings.Join(items, ", "))
	if verbose || Debug {
		fmt.Printf(" %s\n", q)
	}
	_, err = db.Exec(q)
	return err
}

func migrateTable(db dbIface, tableName string, resource Resource, verbose bool) error {
	if verbose {
		fmt.Printf("Migrating table '%s'\n", tableName)
	}
	tableDescription, err := getTableDescription(db, tableName)
	if err != nil {
		return err
	}

	tableDescriptionMap := map[string]bool{}
	for _, item := range tableDescription {
		tableDescriptionMap[item.Field] = true
	}

	items := []string{}

	for _, v := range resource.fieldArrays {
		if !tableDescriptionMap[v.ColumnName] {
			items = append(items, fmt.Sprintf("ADD COLUMN %s", v.fieldDescriptionMysql(resource.fieldTypes)))
		} else {
			tableDescriptionMap[v.ColumnName] = false
		}
	}

	if verbose {
		unusedFields := []string{}
		for k, v := range tableDescriptionMap {
			if v == true {
				unusedFields = append(unusedFields, k)
			}
		}
		if len(unusedFields) > 0 {
			fmt.Printf(" unused fields in model: %s\n", strings.Join(unusedFields, ", "))
		}
	}

	if len(items) == 0 {
		return nil
	}

	q := fmt.Sprintf("ALTER TABLE %s %s;", tableName, strings.Join(items, ", "))
	if verbose || Debug {
		fmt.Printf(" %s\n", q)
	}
	_, err = db.Exec(q)

	return err
}

func getTableDescription(db dbIface, tableName string) (map[string]*mysqlColumn, error) {
	columns := map[string]*mysqlColumn{}
	rows, err := db.Query(fmt.Sprintf("describe `%s`;", tableName))
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

func (app *App) unsafeDropTables() error {
	for _, resource := range app.resources {
		err := resource.unsafeDropTable()
		if err != nil {
			return err
		}
	}
	return nil
}
