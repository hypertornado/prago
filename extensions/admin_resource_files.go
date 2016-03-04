package extensions

import (
	"bytes"
	"errors"
	"github.com/hypertornado/prago"
	"github.com/nfnt/resize"
	"github.com/renstrom/shortuuid"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
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

func ResizeImage(param, id string) (out []byte, err error) {
	wrongFormat := errors.New("Wrong format of params")

	id = strings.Split(id, ".")[0]

	var file *os.File
	file, err = os.Open("public/files/uploaded/" + id + ".jpg")
	if err != nil {
		return
	}

	params := strings.Split(param, "x")
	if len(params) != 2 {
		err = wrongFormat
		return
	}

	uintParams := []uint{0, 0}

	for k, v := range params {
		var intVal int
		intVal, err = strconv.Atoi(v)
		if err != nil {
			return
		}
		if intVal < 0 {
			err = wrongFormat
			return
		}
		uintParams[k] = uint(intVal)
	}

	var img image.Image
	img, err = jpeg.Decode(file)
	if err != nil {
		return
	}
	file.Close()

	img = resize.Resize(uintParams[0], uintParams[1], img, resize.Lanczos3)
	buf := bytes.NewBufferString("")
	jpeg.Encode(buf, img, nil)
	return ioutil.ReadAll(buf)
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
		BindData(item, request, BindDataFilterDefault)

		/*fr, ok := item.(*FileResource)
		if !ok {
			panic("wrong type")
		}*/

		/*err = NewImageFromMultipartForm(request.Request(), "file")
		if err != nil {
			panic(err)
		}*/

		err = resource.Create(item)
		if err != nil {
			panic(err)
		}

		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
	return nil
}

func NewImage(data io.ReadCloser, fileType string) (uuid string, err error) {
	defer data.Close()

	img, err := jpeg.Decode(data)
	if err != nil {
		return "", err
	}

	uuid = shortuuid.UUID()

	file, err := os.Create("public/files/uploaded/" + uuid + "." + fileType)
	if err != nil {
		return
	}

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return
	}
	return
}

func NewImageFromMultipartForm(request *http.Request, name string) (string, error) {
	files := request.MultipartForm.File[name]

	if len(files) != 1 {
		return "", errors.New("not one image specified")
	}

	file := files[0]
	fileType := ""
	if strings.HasSuffix(file.Filename, ".jpg") || strings.HasSuffix(file.Filename, ".jpeg") {
		fileType = "jpg"
	}

	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	return NewImage(f, fileType)
}
