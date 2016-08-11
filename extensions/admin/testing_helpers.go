package admin

import (
	"github.com/hypertornado/prago/extensions"
)

func NewAdminMockup(user, password, dbName string) (*Admin, error) {
	db, err := extensions.ConnectMysql(user, password, dbName)
	if err != nil {
		return nil, err
	}

	admin := NewAdmin("test", "test")
	admin.db = db

	return admin, nil
}
