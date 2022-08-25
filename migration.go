package prago

import (
	"database/sql"
	"fmt"
	"strings"
)

func (app *App) initMigrationCommand() {
	app.addCommand("migratedb").Description("migrate database").
		Callback(func() {
			app.Log().Println("Migrating database")
			err := app.migrate(true)
			if err == nil {
				app.Log().Println("Migration done")
			} else {
				panic(err)
			}
		})
}

func (app *App) migrate(verbose bool) error {
	tables, err := listTables(app.db)
	if err != nil {
		return err
	}
	for _, resourceData := range app.resources {
		tables[resourceData.getID()] = false
		err := resourceData.migrate(verbose)
		if err != nil {
			return err
		}
	}

	if verbose {
		unusedTables := []string{}
		for k, v := range tables {
			if v {
				unusedTables = append(unusedTables, k)
			}
		}
		if len(unusedTables) > 0 {
			fmt.Printf("Unused tables: %s\n", strings.Join(unusedTables, ", "))
		}
	}

	return nil
}

func (resourceData *resourceData) unsafeDropTable() error {
	_, err := resourceData.app.db.Exec(fmt.Sprintf("drop table `%s`;", resourceData.id))
	return err
}

func (resourceData *resourceData) migrate(verbose bool) error {
	_, err := getTableDescription(resourceData.app.db, resourceData.id)
	if err == nil {
		return resourceData.migrateTable(resourceData.app.db, resourceData.id, verbose)
	}
	return resourceData.createTable(resourceData.app.db, resourceData.id, verbose)
}

func listTables(db dbIface) (ret map[string]bool, err error) {
	ret = make(map[string]bool)
	var rows *sql.Rows
	rows, err = db.Query("show tables;")
	if err != nil {
		return ret, err
	}
	defer rows.Close()

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

func (resourceData *resourceData) createTable(db dbIface, tableName string, verbose bool) (err error) {
	if verbose {
		fmt.Printf("Creating table '%s'\n", tableName)
	}
	items := []string{}
	for _, v := range resourceData.fields {
		items = append(items, v.fieldDescriptionMysql(resourceData.app.fieldTypes))
	}
	q := fmt.Sprintf("CREATE TABLE %s (%s);", tableName, strings.Join(items, ", "))
	if verbose {
		fmt.Printf(" %s\n", q)
	}
	_, err = db.Exec(q)
	return err
}

// TODO: migrate after each resource initialization
func (resourceData *resourceData) migrateTable(db dbIface, tableName string, verbose bool) error {
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

	for _, v := range resourceData.fields {
		if !tableDescriptionMap[v.id] {
			items = append(items, fmt.Sprintf("ADD COLUMN %s", v.fieldDescriptionMysql(resourceData.app.fieldTypes)))
		} else {
			tableDescriptionMap[v.id] = false
		}
	}

	if verbose {
		unusedFields := []string{}
		for k, v := range tableDescriptionMap {
			if v {
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
	if verbose {
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
	for _, resourceData := range app.resources {
		err := resourceData.unsafeDropTable()
		if err != nil {
			return err
		}
	}
	return nil
}
