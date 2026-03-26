package prago

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
)

func (app *App) uploadFileReader(ctx context.Context, reader io.Reader, fileName string, userID int64, description string) (*File, error) {
	file := File{}
	file.Name = fileName

	uploadData, err := filesCDN.UploadFile(io.NopCloser(reader), file.getExtension())
	if err != nil {
		return nil, fmt.Errorf("uploading file: %s", err)
	}

	file.Width = uploadData.Width
	file.Height = uploadData.Height
	file.UID = uploadData.UUID
	file.User = userID
	file.Description = description

	err = CreateItemWithContext(ctx, app, &file)
	if err != nil {
		return nil, fmt.Errorf("saving file: %s", err)
	}

	return &file, nil
}

func (app *App) UploadFileFromURL(ctx context.Context, fileURL string, userID int64, description string) (*File, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("fetching url: %s", err)
	}
	defer resp.Body.Close()

	fileName := prettyFilename(path.Base(fileURL))
	return app.uploadFileReader(ctx, resp.Body, fileName, userID, description)
}

func (app *App) UploadFile(fileHeader *multipart.FileHeader, request *Request, description string) (*File, error) {
	openedFile, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("opening multipart file: %s", err)
	}
	defer openedFile.Close()

	return app.uploadFileReader(
		request.Request().Context(),
		openedFile,
		prettyFilename(fileHeader.Filename),
		request.UserID(),
		description,
	)
}
