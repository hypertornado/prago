package prago

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hypertornado/prago/pragocdn/cdnclient"
	"github.com/hypertornado/prago/utils"
)

var filesCDN cdnclient.CDNAccount

func initCDN(app *App) {
	cdnURL := app.ConfigurationGetStringWithFallback("cdnURL", "https://www.prago-cdn.com")
	cdnAccount := app.ConfigurationGetStringWithFallback("cdnAccount", app.codeName)
	cdnPassword := app.ConfigurationGetStringWithFallback("cdnPassword", "")
	filesCDN = cdnclient.NewCDNAccount(cdnURL, cdnAccount, cdnPassword)
}

func (app *App) thumb(ids string) string {
	for _, v := range strings.Split(ids, ",") {
		var image File
		err := app.Query().WhereIs("uid", v).Get(&image)
		if err == nil && image.isImage() {
			return image.GetSmall()
		}
	}
	return ""
}

//GetFiles gets files from app
func (app *App) GetFiles(ids string) []*File {
	var files []*File
	idsAr := strings.Split(ids, ",")
	for _, v := range idsAr {
		var image File
		err := app.Query().WhereIs("uid", v).Get(&image)
		if err == nil {
			files = append(files, &image)
		}
	}
	return files
}

//File is structure representing files in admin
type File struct {
	ID          int64  `prago-order-desc:"true" prago-preview:"true"`
	UID         string `prago-unique:"true" prago-preview:"true" prago-type:"cdnfile" prago-description:"File"`
	Name        string `prago-edit:"_"`
	Description string `prago-type:"text" prago-preview:"true"`
	User        int64  `prago-type:"relation" prago-edit:"_"`
	Width       int64  `prago-edit:"_"`
	Height      int64  `prago-edit:"_"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

//UploadFile uploads files to app
func (app *App) UploadFile(fileHeader *multipart.FileHeader, user *User, description string) (*File, error) {
	fileName := utils.PrettyFilename(fileHeader.Filename)
	file := File{}
	file.Name = fileName

	openedFile, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("opening multipart file: %s", err)
	}
	defer openedFile.Close()

	uploadData, err := filesCDN.UploadFile(openedFile, file.GetExtension())
	if err != nil {
		return nil, fmt.Errorf("uploading multipart file: %s", err)
	}

	file.Width = uploadData.Width
	file.Height = uploadData.Height

	file.UID = uploadData.UUID

	if user != nil {
		file.User = user.ID
	}
	file.Description = description
	err = app.Create(&file)
	if err != nil {
		return nil, fmt.Errorf("saving file: %s", err)
	}

	return &file, nil
}

//UpdateMetadata updates metadata of file
func (f *File) updateMetadata() error {
	metadata, err := filesCDN.GetMetadata(f.UID)
	if err != nil {
		return err
	}

	f.Width = metadata.Width
	f.Height = metadata.Height
	return nil
}

//GetExtension gets file extension
func (f File) GetExtension() string {
	extension := filepath.Ext(f.Name)
	extension = strings.Replace(extension, ".", "", -1)
	return extension
}

func getOldRedirectParams(request Request, app *App) (uuid, name string, err error) {
	name = request.Params().Get("name")
	uuid = fmt.Sprintf("%s%s%s%s%s%s",
		request.Params().Get("a"),
		request.Params().Get("b"),
		request.Params().Get("c"),
		request.Params().Get("d"),
		request.Params().Get("e"),
		strings.Split(name, "-")[0],
	)

	var file File
	err = app.Query().WhereIs("uid", uuid).Get(&file)
	if err != nil {
		return
	}
	name = file.Name
	return
}

func initFilesResource(resource *Resource) {
	app := resource.app
	initCDN(app)
	resource.name = messages.GetNameFunction("admin_files")

	resource.fieldMap["uid"].HumanName = messages.GetNameFunction("admin_file")
	resource.fieldMap["width"].HumanName = messages.GetNameFunction("width")
	resource.fieldMap["height"].HumanName = messages.GetNameFunction("height")

	app.AddCommand("files", "metadata").
		Callback(func() {
			var files []*File
			must(app.Query().Get(&files))
			for _, v := range files {
				err := v.updateMetadata()
				if err != nil {
					fmt.Println("error while updating metadata: ", v.ID, err)
					continue
				}
				file := *v
				err = app.Save(&file)
				if err != nil {
					fmt.Println("error while saving file: ", v.ID)
				} else {
					fmt.Println("saved ok: ", v.ID, v.Width, v.Height)
				}
			}
		})

	resource.resourceController.addBeforeAction(func(request Request) {
		if request.Request().Method == "POST" && strings.HasSuffix(request.Request().URL.Path, "/delete") {
			idStr := request.Params().Get("id")
			id, err := strconv.Atoi(idStr)
			if err == nil {
				var file File
				must(app.Query().WhereIs("id", id).Get(&file))
				err = filesCDN.DeleteFile(file.UID)
				if err != nil {
					app.Log().Printf("deleting CDN: %s\n", err)
				}
			}
		}
	})

	app.mainController.get("/files/thumb/:size/:a/:b/:c/:d/:e/:name", func(request Request) {
		uuid, name, err := getOldRedirectParams(request, app)
		if err != nil {
			panic(err)
		}

		var size string
		switch request.Params().Get("size") {
		case "large":
			size = "1000"
		case "medium":
			size = "400"
		case "small":
			size = "200"
		default:
			panic("wrong size")
		}

		request.Redirect(filesCDN.GetImageURL(uuid, name, size))
	}, func(params map[string]string) bool {
		size := params["size"]
		if size == "large" || size == "medium" || size == "small" {
			return true
		}
		return false
	})

	app.mainController.get("/files/original/:a/:b/:c/:d/:e/:name", func(request Request) {
		uuid, name, err := getOldRedirectParams(request, app)
		must(err)
		request.Redirect(filesCDN.GetFileURL(uuid, name))
	})

	resource.defaultItemsPerPage = 100

	//TODO: authorize
	resource.resourceController.post(resource.getURL(""), func(request Request) {
		validateCSRF(request)

		multipartFiles := request.Request().MultipartForm.File["uid"]
		if len(multipartFiles) != 1 {
			panic("must have 1 file selected")
		}

		user := request.GetUser()

		_, err := resource.app.UploadFile(multipartFiles[0], &user, request.Params().Get("Description"))
		must(err)
		request.AddFlashMessage(messages.Get(getLocale(request), "admin_item_created"))
		request.Redirect(resource.getURL(""))
	})

	resource.Action("getcdnurl").Method("POST").Handler(
		func(request Request) {
			uuid := request.Params().Get("uuid")
			size := request.Params().Get("size")

			files := resource.app.GetFiles(uuid)
			if len(files) == 0 {
				panic("can't find file")
			}
			file := files[0]

			redirectURL := filesCDN.GetImageURL(uuid, file.Name, size)
			request.Redirect(redirectURL)
		},
	)
}

func (f *File) getPath(prefix string) (folder, file string) {
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

type imageResponse struct {
	ID          int64
	UID         string
	Name        string
	Description string
	Thumb       string
}

func writeFileResponse(request Request, files []*File) {
	responseData := []*imageResponse{}
	for _, v := range files {
		ir := &imageResponse{
			ID:          v.ID,
			UID:         v.UID,
			Name:        v.Name,
			Description: v.Description,
		}

		ir.Thumb = v.GetMedium()

		responseData = append(responseData, ir)
	}
	request.RenderJSON(responseData)
}

//GetLarge file path
func (f *File) GetLarge() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "1000")
}

//GetGiant file path
func (f *File) GetGiant() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "2500")
}

//GetMedium file path
func (f *File) GetMedium() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "400")
}

//GetSmall file path
func (f *File) GetSmall() string {
	return filesCDN.GetImageURL(f.UID, f.Name, "200")
}

//GetOriginal file path
func (f *File) GetOriginal() string {
	return filesCDN.GetFileURL(f.UID, f.Name)
}

//GetMetadataPath gets metadada file path
func (f *File) GetMetadataPath() string {
	return filesCDN.MetadataPath(f.UID)
}

//IsImage detects if file is image
func (f *File) IsImage() bool {
	return f.isImage()
}

func (f *File) isImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
		return true
	}
	return false
}
