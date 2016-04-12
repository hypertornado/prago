package admin

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
	"mime/multipart"
	"os"
	"strconv"
	"strings"
)

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
