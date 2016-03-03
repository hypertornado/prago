package extensions

import (
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	//"github.com/nfnt/resize"
	"github.com/renstrom/shortuuid"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type FileResource struct {
	ID           int64
	Name         string
	UUID         string `prago-admin-access:"-" prago-admin-type:"image" prago-admin-show:"yes"`
	Description  string `prago-admin-type:"text"`
	Width        int64
	Height       int64
	OriginalName string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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

func ResizeImage(params, id string) (bytes []byte, err error) {
	id = strings.Split(id, ".")[0]

	var file *os.File
	file, err = os.Open("public/img/uploaded/" + id + ".jpg")
	if err != nil {
		return
	}

	fmt.Println(file)

	return ioutil.ReadAll(file)
}

func (FileResource) AdminInitResource(a *Admin, resource *AdminResource) error {
	BindList(a, resource)
	BindNew(a, resource)
	BindDetail(a, resource)
	BindUpdate(a, resource)
	BindDelete(a, resource)

	resource.ResourceController.Get("/img/:resize/:id", func(request prago.Request) {
		bytes, err := ResizeImage(request.Params().Get("resize"), request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		request.Response().WriteHeader(200)

		_, err = request.Response().Write(bytes)
		if err != nil {
			panic(err)
		}

		request.SetProcessed()

	})

	resource.ResourceController.Post(a.GetURL(resource, ""), func(request prago.Request) {
		item, err := resource.NewItem()
		if err != nil {
			panic(err)
		}
		BindData(item, request.Params(), BindDataFilterDefault)

		fr, ok := item.(*FileResource)
		if !ok {
			panic("wrong type")
		}

		err = fr.NewImageFromMultipartForm(request.Request(), "file")
		if err != nil {
			panic(err)
		}

		err = resource.Create(item)
		if err != nil {
			panic(err)
		}

		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
	return nil
}

func (fr *FileResource) NewImage(data io.ReadCloser, fileType string) (err error) {
	defer data.Close()

	img, err := jpeg.Decode(data)
	if err != nil {
		return err
	}

	uuid := shortuuid.UUID()

	file, err := os.Create("public/img/uploaded/" + uuid + "." + fileType)
	if err != nil {
		return
	}

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return
	}

	fr.UUID = uuid
	fr.Width = int64(img.Bounds().Max.X)
	fr.Height = int64(img.Bounds().Max.Y)

	return
}

func (fr *FileResource) NewImageFromMultipartForm(request *http.Request, name string) error {
	files := request.MultipartForm.File[name]
	if len(files) != 1 {
		return errors.New("not one image specified")
	}

	file := files[0]
	fileType := ""
	if strings.HasSuffix(file.Filename, ".jpg") || strings.HasSuffix(file.Filename, ".jpeg") {
		fileType = "jpg"
	}

	f, err := file.Open()
	if err != nil {
		return err
	}

	return fr.NewImage(f, fileType)
}
