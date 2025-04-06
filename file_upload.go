package prago

import (
	"fmt"
	"mime/multipart"
)

func (app *App) UploadFile(fileHeader *multipart.FileHeader, request *Request, description string) (*File, error) {
	fileName := prettyFilename(fileHeader.Filename)
	file := File{}
	file.Name = fileName

	openedFile, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("opening multipart file: %s", err)
	}
	defer openedFile.Close()

	uploadData, err := filesCDN.UploadFile(openedFile, file.getExtension())
	if err != nil {
		return nil, fmt.Errorf("uploading multipart file: %s", err)
	}

	file.Width = uploadData.Width
	file.Height = uploadData.Height

	file.UID = uploadData.UUID

	file.User = request.UserID()
	file.Description = description
	err = CreateItemWithContext(request.Request().Context(), app, &file)
	if err != nil {
		return nil, fmt.Errorf("saving file: %s", err)
	}

	return &file, nil
}
