package extensions

/*import (
	"errors"
	"github.com/renstrom/shortuuid"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"strings"
)

type Image struct {
	ID          int64  `sql:"not null;unique"`
	Uuid        string `sql:"not null;unique"`
	Description string
	Width       int
	Height      int
	Filetype    string
}

func NewImage(organization Organization, data io.ReadCloser, fileType string) (image *Image, err error) {
	defer data.Close()

	img, err := jpeg.Decode(data)
	if err != nil {
		return nil, err
	}

	image = &Image{
		Uuid:     shortuuid.UUID(),
		Width:    img.Bounds().Max.X,
		Height:   img.Bounds().Max.Y,
		Filetype: fileType,
	}

	file, err := os.Create("public/images/" + image.Uuid + "." + fileType)
	if err != nil {
		return
	}

	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return
	}

	//err = model.db.Save(image).Error
	return
}

func (model *Model) NewImageFromMultipartForm(organization Organization, request *http.Request, name string) (image *Image, err error) {
	files := request.MultipartForm.File[name]
	if len(files) != 1 {
		return nil, errors.New("not one image specified")
	}

	file := files[0]
	fileType := ""
	if strings.HasSuffix(file.Filename, ".jpg") || strings.HasSuffix(file.Filename, ".jpeg") {
		fileType = "jpg"
	}

	f, err := file.Open()
	if err != nil {
		return nil, err
	}

	return model.NewImage(organization, f, fileType)
}

func (model *Model) ListOrganizationImages(organization Organization) (images []Image, err error) {
	err = model.db.Where(&Image{OrganizationId: organization.Id}).Find(&images).Error
	return
}*/
