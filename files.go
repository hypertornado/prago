package prago

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hypertornado/prago/messages"
	"github.com/hypertornado/prago/pragocdn/cdnclient"
	"github.com/hypertornado/prago/utils"
)

var filesCDN cdnclient.CDNAccount

func initCDN(a *Administration) {
	cdnURL := a.App.Config.GetStringWithFallback("cdnURL", "https://www.prago-cdn.com")
	cdnAccount := a.App.Config.GetStringWithFallback("cdnAccount", a.HumanName)
	cdnPassword := a.App.Config.GetStringWithFallback("cdnPassword", "")
	filesCDN = cdnclient.NewCDNAccount(cdnURL, cdnAccount, cdnPassword)
}

func (admin *Administration) thumb(ids string) string {
	for _, v := range strings.Split(ids, ",") {
		var image File
		err := admin.Query().WhereIs("uid", v).Get(&image)
		if err == nil && image.IsImage() {
			return image.GetSmall()
		}
	}
	return ""
}

func (admin *Administration) GetFiles(ids string) []*File {
	var files []*File
	idsAr := strings.Split(ids, ",")
	for _, v := range idsAr {
		var image File
		err := admin.Query().WhereIs("uid", v).Get(&image)
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

func (admin *Administration) UploadFile(fileHeader *multipart.FileHeader, user *User, description string) (*File, error) {
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
	err = admin.Create(&file)
	if err != nil {
		return nil, fmt.Errorf("saving file: %s", err)
	}

	return &file, nil
}

func (f *File) UpdateMetadata() error {
	metadata, err := filesCDN.GetMetadata(f.UID)
	if err != nil {
		return err
	}

	f.Width = metadata.Width
	f.Height = metadata.Height
	return nil
}

func (f File) GetExtension() string {
	extension := filepath.Ext(f.Name)
	extension = strings.Replace(extension, ".", "", -1)
	return extension
}

func getOldRedirectParams(request Request, admin *Administration) (uuid, name string, err error) {
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
	err = admin.Query().WhereIs("uid", uuid).Get(&file)
	if err != nil {
		return
	}
	name = file.Name
	return
}

func initFilesResource(resource *Resource) {
	a := resource.Admin
	initCDN(a)
	resource.HumanName = messages.Messages.GetNameFunction("admin_files")
	app := resource.Admin.App

	resource.fieldMap["uid"].HumanName = messages.Messages.GetNameFunction("admin_file")
	resource.fieldMap["width"].HumanName = messages.Messages.GetNameFunction("width")
	resource.fieldMap["height"].HumanName = messages.Messages.GetNameFunction("height")

	app.AddCommand("files", "metadata").
		Callback(func() {
			var files []*File
			must(a.Query().Get(&files))
			for _, v := range files {
				err := v.UpdateMetadata()
				if err != nil {
					fmt.Println("error while updating metadata: ", v.ID, err)
					continue
				}
				file := *v
				err = a.Save(&file)
				if err != nil {
					fmt.Println("error while saving file: ", v.ID)
				} else {
					fmt.Println("saved ok: ", v.ID, v.Width, v.Height)
				}
			}
		})

	resource.ResourceController.AddBeforeAction(func(request Request) {
		if request.Request().Method == "POST" && strings.HasSuffix(request.Request().URL.Path, "/delete") {
			idStr := request.Params().Get("id")
			id, err := strconv.Atoi(idStr)
			if err == nil {
				var file File
				must(a.Query().WhereIs("id", id).Get(&file))
				err = filesCDN.DeleteFile(file.UID)
				if err != nil {
					a.App.Log().Printf("deleting CDN: %s\n", err)
				}
			}
		}
	})

	a.App.MainController().Get("/files/thumb/:size/:a/:b/:c/:d/:e/:name", func(request Request) {
		uuid, name, err := getOldRedirectParams(request, a)
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

	a.App.MainController().Get("/files/original/:a/:b/:c/:d/:e/:name", func(request Request) {
		uuid, name, err := getOldRedirectParams(request, a)
		must(err)
		request.Redirect(filesCDN.GetFileURL(uuid, name))
	})

	resource.ItemsPerPage = 100

	bindImageAPI(a)

	//TODO: authorize
	resource.ResourceController.Post(resource.GetURL(""), func(request Request) {
		ValidateCSRF(request)

		multipartFiles := request.Request().MultipartForm.File["uid"]
		if len(multipartFiles) != 1 {
			panic("must have 1 file selected")
		}

		user := GetUser(request)

		_, err := resource.Admin.UploadFile(multipartFiles[0], &user, request.Params().Get("Description"))
		must(err)
		AddFlashMessage(request, messages.Messages.Get(getLocale(request), "admin_item_created"))
		request.Redirect(resource.GetURL(""))
	})

	resource.AddAction(Action{
		Method: "POST",
		URL:    "getcdnurl",
		Handler: func(resource Resource, request Request, user User) {
			uuid := request.Params().Get("uuid")
			size := request.Params().Get("size")

			files := resource.Admin.GetFiles(uuid)
			if len(files) == 0 {
				panic("can't find file")
			}
			file := files[0]

			redirectURL := filesCDN.GetImageURL(uuid, file.Name, size)
			request.Redirect(redirectURL)
		},
	})
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

		//v.GetMedium()

		//_, fileURL := v.getPath(fileDownloadPath + "thumb/small")
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

func (f *File) GetMetadataPath() string {
	return filesCDN.MetadataPath(f.UID)
}

func (f *File) IsImage() bool {
	if strings.HasSuffix(f.Name, ".jpg") || strings.HasSuffix(f.Name, ".jpeg") || strings.HasSuffix(f.Name, ".png") {
		return true
	}
	return false
}
