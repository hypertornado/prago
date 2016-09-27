package admin

import (
	"database/sql"
)

type Transaction struct {
	tx    *sql.Tx
	admin *Admin
	err   error
}

func (a *Admin) Transaction() (transaction *Transaction) {
	tx, err := a.getDB().Begin()
	transaction = &Transaction{
		err: err,
	}
	if err != nil {
		return
	}

	transaction.tx = tx
	transaction.admin = a
	return
}

func (t *Transaction) Create(item interface{}) error {
	resource, err := t.admin.getResourceByItem(item)
	if err != nil {
		return err
	}

	return resource.createWithDBIface(item, t.tx)
}

func (t *Transaction) Save(item interface{}) error {
	resource, err := t.admin.getResourceByItem(item)
	if err != nil {
		return err
	}

	return resource.saveWithDBIface(item, t.tx)
}

func (t *Transaction) Query() *Query {
	if t.err != nil {
		return &Query{err: t.err}
	}
	return &Query{
		query: &listQuery{},
		admin: t.admin,
		db:    t.tx,
	}
}

func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *Transaction) Tx() *sql.Tx {
	return t.tx
}
