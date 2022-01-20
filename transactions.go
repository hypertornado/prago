package prago

//TODO: fix transactions later
/*
//Transaction represents sql transaction
type Transaction struct {
	tx  *sql.Tx
	app *App
	err error
	//resource *resource
}

func (app *App) Transaction() (t *Transaction) {
	tx, err := app.db.Begin()
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

	return resource.createWithDBIface(item, t.tx, false)
}

//Save transaction
func (t *Transaction) Save(item interface{}) error {
	resource, err := t.app.getResourceByItem(item)
	if err != nil {
		return err
	}

	return resource.saveWithDBIface(item, t.tx, false)
}

//Query with transaction
func TransactionQuery[T any](app *App, t *Transaction) *Query[T] {
	resource := GetResource[T](app)
	ret := resource.Query()

	ret.query.db = t.tx
	return ret
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
*/
