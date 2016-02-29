package extensions

import (
	"time"
)

type FileResource struct {
	ID          int64
	Name        string
	UID         string `prago-admin-access:"-"`
	Description string `prago-admin-type:"text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (FileResource) AdminName() string { return "Soubory" }
func (FileResource) AdminID() string   { return "files" }

func (FileResource) GetFormItems(ar *AdminResource, item interface{}) ([]AdminFormItem, error) {
	items, err := GetFormItemsDefault(ar, item)

	newItem := AdminFormItem{
		Name:      "file",
		NameHuman: "File",
		Template:  "admin_item_file",
	}

	items = append([]AdminFormItem{newItem}, items...)

	return items, err
}
