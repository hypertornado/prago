package extensions

import (
	"github.com/hypertornado/prago"
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

func (FileResource) AdminInitResource(a *Admin, resource *AdminResource) error {
	BindList(a, resource)
	BindNew(a, resource)
	BindDetail(a, resource)
	BindUpdate(a, resource)
	BindDelete(a, resource)

	resource.ResourceController.Post(resource.ResourceURL(""), func(request prago.Request) {
		/*err := resource.CreateItemFromParams(request.Params())
		if err != nil {
			panic(err)
		}
		prago.Redirect(request, a.Prefix+"/"+resource.ID)*/

		panic("NNN")

	})

	return nil
}
