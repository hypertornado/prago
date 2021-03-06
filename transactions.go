package prago

import (
	"database/sql"
)

//Transaction represents sql transaction
type Transaction struct {
	tx  *sql.Tx
	app *App
	err error
}

//Transaction creates transaction
func (app *App) Transaction() (t *Transaction) {
	tx, err := app.getDB().Begin()
	t = &Transaction{
		err: err,
	}
	if err != nil {
		return
	}

	t.tx = tx
	t.app = app
	return
}

//Create transaction
func (t *Transaction) Create(item interface{}) error {
	resource, err := t.app.getResourceByItem(item)
	if err != nil {
		return err
	}

	return resource.createWithDBIface(item, t.tx)
}

//Save transaction
func (t *Transaction) Save(item interface{}) error {
	resource, err := t.app.getResourceByItem(item)
	if err != nil {
		return err
	}

	return resource.saveWithDBIface(item, t.tx)
}

//Query with transaction
func (t *Transaction) Query() *Query {
	if t.err != nil {
		return &Query{err: t.err}
	}
	return &Query{
		query: &listQuery{},
		app:   t.app,
		db:    t.tx,
	}
}

//Commit transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

//Rollback transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

//Tx returns raw transaction
func (t *Transaction) Tx() *sql.Tx {
	return t.tx
}
