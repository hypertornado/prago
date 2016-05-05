package admin

import (
	"bytes"
	"errors"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"github.com/hypertornado/prago/utils"
	"github.com/nfnt/resize"
	"github.com/renstrom/shortuuid"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"
)

type File struct {
	ID          int64
	Name        string
	Description string `prago-type:"text"`
	UID         string `prago-unique:"true"`
	Extension   string
	Size        int64
	Width       int64
	Height      int64
	UserId      int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (File) AdminName(lang string) string { return messages.Messages.Get(lang, "admin_files") }

func (File) AdminAfterFormCreated(f *Form, request prago.Request, newItem bool) *Form {
	newForm := NewForm()
	newForm.Method = f.Method
	newForm.Action = f.Action
	if newItem {
		newForm.AddFileInput("file", messages.Messages.Get(GetLocale(request), "admin_file"))
		newForm.AddTextareaInput("Description", messages.Messages.Get(GetLocale(request), "Description"))
		newForm.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_create"))
	} else {
		fi := newForm.AddTextInput("Name", messages.Messages.Get(GetLocale(request), "Name"))
		fi.Readonly = true
		fi.Value = f.GetItemByName("Name").Value

		fi = newForm.AddTextInput("url", messages.Messages.Get(GetLocale(request), "Url"))
		fi.Readonly = true
		fi.Value = f.GetItemByName("UID").Value

		fi = newForm.AddTextareaInput("Description", messages.Messages.Get(GetLocale(request), "Description"))
		fi.Value = f.GetItemByName("Description").Value
		fi.Focused = true
		newForm.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_edit"))
	}
	AddCSRFToken(newForm, request)
	return newForm
}

func (File) AdminInitResource(a *Admin, resource *AdminResource) error {

	resource.ResourceController.Post(a.GetURL(resource, ""), func(request prago.Request) {
		ValidateCSRF(request)

		multipartFiles := request.Request().MultipartForm.File["file"]
		if len(multipartFiles) != 1 {
			panic("must have 1 file selected")
		}

		fileName := utils.PrettyFilename(multipartFiles[0].Filename)

		item, err := resource.NewItem()
		prago.Must(err)
		file := item.(*File)
		file.Name = fileName
		file.Description = request.Params().Get("Description")
		file.UID = shortuuid.UUID()

		folderPath, filePath := file.GetPath("public/xfiles")
		prago.Must(loadFile(folderPath, filePath, multipartFiles[0]))

		prago.Must(resource.Create(file))
		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})

	resource.Actions["create"] = nil

	return nil
}

func (f *File) SaveData(header *multipart.FileHeader) error {
	//header.Open()
	return nil
}

func (f *File) GetPath(prefix string) (folder, file string) {
	pathSeparator := "/"
	folder = prefix
	if len(folder) > 0 && !strings.HasSuffix(folder, pathSeparator) {
		folder += pathSeparator
	}

	if len(f.UID) < 7 {
		panic("too short uid")
	}

	uidPrefix := f.UID[0:5]
	folders := strings.Split(uidPrefix, "")
	folder += strings.Join(folders, pathSeparator)

	file = folder + pathSeparator + f.UID[5:] + "-" + f.Name
	return
}

func loadFile(folder, path string, header *multipart.FileHeader) error {
	if !strings.HasPrefix(path, folder) {
		return errors.New("folder path should be prefix of path")
	}
	err := os.MkdirAll(folder, 0777)
	if err != nil {
		return err
	}

	inFile, err := header.Open()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}

	io.Copy(outFile, inFile)

	defer outFile.Close()
	defer inFile.Close()
	return nil
}

var FilesBasePath = ""

func ResizeImage(param, id string) (out []byte, err error) {
	wrongFormat := errors.New("Wrong format of params")

	id = strings.Split(id, ".")[0]

	var file *os.File
	file, err = os.Open(FilesBasePath + "/" + id + ".jpg")
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

func BindImageResizer(controller *prago.Controller, path string) {
	FilesBasePath = path
	controller.Get("/img/:resize/:id", func(request prago.Request) {
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
}

func NewImage(data io.ReadCloser, fileType string) (uuid string, err error) {
	defer data.Close()

	img, err := jpeg.Decode(data)
	if err != nil {
		return "", err
	}

	uuid = shortuuid.UUID()

	file, err := os.Create(FilesBasePath + "/" + uuid + "." + fileType)
	if err != nil {
		return
	}

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return
	}
	return
}

func NewImageFromMultipartForm(form *multipart.Form, formItemName string) (string, error) {

	files := form.File[formItemName]

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
