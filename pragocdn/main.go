package main

import (
	"compress/gzip"
	"embed"
	"errors"
	"fmt"
	"image"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	_ "image/jpeg"
	_ "image/png"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/pragocdn/cdnclient"
)

const version = "2022.15"

var app *prago.App

var homePath = os.Getenv("HOME")

var uuidRegex = regexp.MustCompile("^[a-zA-Z0-9]{10,}$")
var filenameRegex = regexp.MustCompile("^[a-zA-Z0-9_.-]{1,150}$")
var extensionRegex = regexp.MustCompile("^[a-zA-Z0-9]{1,10}$")

var cmykProfilePath = os.Getenv("HOME") + "/.pragocdn/cmyk.icm"

var vipsMutexes []*sync.Mutex

var fileExtensionMap = map[string]string{
	"jpeg": "jpg",
}

//go:embed resources/icons/*
var iconsFS embed.FS

func main() {

	for i := 0; i < 10; i++ {
		vipsMutexes = append(vipsMutexes, &sync.Mutex{})
	}

	app = prago.New("pragocdn", version)

	app.SetIcons(iconsFS, "resources/icons/")

	initCDNProjectResource()
	bindStats(app)
	bindCDNFiles(app)

	app.GET("/", func(request *prago.Request) {
		out := fmt.Sprintf("Prago CDN\nhttps://www.prago-cdn.com\nversion %s\nadmin Ondřej Odcházel, https//www.odchazel.com", version)
		http.Error(request.Response(), out, 200)
	})

	app.POST("/:account/upload/:extension", func(request *prago.Request) {
		defer request.Request().Body.Close()
		accountName := request.Param("account")
		project := getCDNProject(accountName)
		if project == nil {
			panic("no account")
		}

		authorization := request.Request().Header.Get("X-Authorization")
		if project.Password != authorization {
			panic("wrong authorization")
		}

		extension := normalizeExtension(request.Param("extension"))
		data, err := project.uploadFile(extension, request.Request().Body)
		if err != nil {
			panic(err)
		}

		request.RenderJSON(data)
	})

	app.GET("/:account/:uuid/metadata", func(request *prago.Request) {
		metadata, err := getMetadata(
			request.Param("account"),
			request.Param("uuid"),
		)
		if err != nil {
			panic(err)
		}
		request.RenderJSON(metadata)
	})

	app.GET("/:account/:uuid/:format/:hash/:name", func(request *prago.Request) {
		errCode, err, stream, mimeExtension, size := getFile(
			request.Param("account"),
			request.Param("uuid"),
			request.Param("format"),
			request.Param("hash"),
			request.Param("name"),
		)
		if stream != nil {
			defer stream.Close()
		}

		if err != nil {
			if app.DevelopmentMode() {
				panic(err)
			} else {
				path := request.Request().URL.Path
				fmt.Printf("getFile error on path %s: %s\n", path, err)
				return
			}
		}

		//https://gist.github.com/the42/1956518
		switch errCode {
		case 404:
			render404(request)
			return
		case 498:
			render498(request)
			return
		}

		if err != nil {
			panic(err)
		}

		request.Response().Header().Set("Cache-Control", "public, max-age=31536000")
		request.Response().Header().Set("X-Content-Type-Options", "nosniff")
		request.Response().Header().Set("Content-Type", mimeExtension)

		if strings.HasPrefix(mimeExtension, "text/") && strings.Contains(request.Request().Header.Get("Accept-Encoding"), "gzip") {
			request.Response().Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(request.Response())
			_, err = io.Copy(gz, stream)
			defer gz.Close()
			if err != nil {
				panic(err)
			}
		} else {
			request.Response().Header().Set("Content-Length", fmt.Sprintf("%d", size))
			_, err = io.Copy(request.Response(), stream)
			if err != nil {
				panic(err)
			}
		}
	})

	app.DELETE("/:account/:uuid", func(request *prago.Request) {
		err := deleteFile(
			request.Param("account"),
			request.Request().Header.Get("X-Authorization"),
			request.Param("uuid"),
		)
		if err != nil {
			panic(err)
		}
		request.RenderJSON(true)
	})

	app.Run()
}

func (account *CDNProject) uploadFile(extension string, inData io.Reader) (*cdnclient.CDNFileData, error) {
	uuid := RandomString(20)
	dirPath := getFileDirectoryPath(account.Name, uuid)

	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return nil, err
	}

	filePath := getFilePath(account.Name, uuid, extension)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, inData)
	if err != nil {
		return nil, err
	}

	cdnFile := account.createFile(uuid, extension)
	cdnFile.update()

	return getMetadata(account.Name, uuid)
}

func deleteFile(accountName, password, uuid string) error {
	project := getCDNProject(accountName)
	if project == nil {
		return errors.New("account not found")
	}

	if !uuidRegex.MatchString(uuid) {
		return errors.New("wrongs uuid format: " + uuid)
	}

	if project.Password != password {
		return errors.New("wrong password")
	}

	filePath, _, err := getFilePathFromUUID(accountName, uuid)
	if err != nil {
		return err
	}

	deletedDir := fmt.Sprintf("%s/.pragocdn/deleted/%s", homePath, accountName)

	cmd := exec.Command("mv", filePath, deletedDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getFilePathFromUUID(accountName, uuid string) (filePath, extension string, err error) {
	dirPath := getFileDirectoryPath(accountName, uuid)
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return "", "", err
	}

	var name string
	for _, v := range files {
		fileName := v.Name()
		if strings.HasPrefix(fileName, uuid+".") {
			_, extension, _ = getNameAndExtension(fileName)
			name = fileName
			break
		}
	}

	if name == "" {
		return "", "", errors.New("no file found for uuid: " + uuid)
	}

	return fmt.Sprintf("%s/%s", dirPath, name), extension, nil
}

// TODO: cache somewhere
func getMetadata(accountName, uuid string) (*cdnclient.CDNFileData, error) {
	filePath, extension, err := getFilePathFromUUID(accountName, uuid)
	if err != nil {
		return nil, fmt.Errorf("getting file path from uuid: %s", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()

	var width, height int

	if isImageExtension(extension) {
		i, _, err := image.Decode(file)
		if err == nil {
			bounds := i.Bounds()
			width = bounds.Max.X
			height = bounds.Max.Y
		} else {
			app.Log().Errorf("decoding: %s", err)
		}
	}

	filestat, _ := file.Stat()

	return &cdnclient.CDNFileData{
		UUID:      uuid,
		Extension: extension,
		IsImage:   isImageExtension(extension),
		Filesize:  filestat.Size(),
		Width:     int64(width),
		Height:    int64(height),
	}, nil

}

func getFile(accountName, uuid, format, hash, name string) (eddCode int, err error, source io.ReadCloser, mimeExtension string, size int64) {
	project := getCDNProject(accountName)
	if project == nil {
		return 404, errors.New("account not found"), nil, "", -1
	}

	if !uuidRegex.MatchString(uuid) {
		return 404, errors.New("wrongs uuid format: " + uuid), nil, "", -1
	}

	_, fileExtension, err := getNameAndExtension(name)
	if err != nil {
		return 404, err, nil, "", -1
	}

	mimeExtension = mime.TypeByExtension("." + fileExtension)

	expectedHash := cdnclient.GetHash(
		project.Name,
		project.Password,
		uuid,
		format,
		name,
	)
	if expectedHash != hash {
		return 498, errors.New("wrong hash"), nil, "", -1
	}

	originalPath := getFilePath(accountName, uuid, fileExtension)

	var path string
	if format == "file" {
		path = originalPath
	} else {
		path, err = convertedFilePath(accountName, uuid, fileExtension, format)
		if err != nil {
			return 404, err, nil, "", -1
		}
		mimeExtension = mime.TypeByExtension(".webp")
	}

	file, err := os.Open(path)
	if err != nil {
		return 500, err, nil, "", -1
	}

	stat, err := file.Stat()
	if err != nil {
		return 500, err, nil, "", -1
	}

	return 200, nil, file, mimeExtension, stat.Size()
}

func getFileDirectoryPath(account, uuid string) string {
	firstPrefix := strings.ToLower(uuid[0:2])
	secondPrefix := strings.ToLower(uuid[2:4])
	return fmt.Sprintf("%s/.pragocdn/files/%s/%s/%s",
		homePath,
		account,
		firstPrefix,
		secondPrefix,
	)
}

func getFilePath(account, uuid, extension string) string {
	return fmt.Sprintf("%s/%s.%s",
		getFileDirectoryPath(account, uuid),
		uuid,
		extension,
	)
}

func getCacheDirectoryPath(account, uuid, format string) string {
	firstPrefix := strings.ToLower(uuid[0:2])
	secondPrefix := strings.ToLower(uuid[2:4])
	return fmt.Sprintf("%s/.pragocdn/cache/%s/%s/%s/%s",
		homePath,
		account,
		format,
		firstPrefix,
		secondPrefix,
	)
}

func getCacheFilePath(account, uuid, format, extension string) string {
	return fmt.Sprintf("%s/%s.%s",
		getCacheDirectoryPath(account, uuid, format),
		uuid,
		extension,
	)
}

var singleSizeRegexp = regexp.MustCompile("^[1-9][0-9]*$")
var sizeRegexp = regexp.MustCompile("^[1-9][0-9]{0,3}x[1-9][0-9]{0,3}$")

func isImageExtension(extension string) bool {
	if extension == "jpg" || extension == "png" || extension == "webp" {
		return true
	}
	return false
}

func convertedFilePath(account, uuid, extension, format string) (string, error) {
	if !isImageExtension(extension) {
		return "", errors.New("cant resize non images")
	}

	originalPath := getFilePath(account, uuid, extension)
	outputFilePath := getCacheFilePath(account, uuid, format, "webp")
	outputDirectoryPath := getCacheDirectoryPath(account, uuid, format)

	if singleSizeRegexp.MatchString(format) {
		return outputFilePath, vipsThumbnail(originalPath, outputDirectoryPath, outputFilePath, format, false)
	}

	if sizeRegexp.MatchString(format) {
		return outputFilePath, vipsThumbnail(originalPath, outputDirectoryPath, outputFilePath, format, true)
	}

	return "", errors.New("wrong file convert format")
}

func render404(request *prago.Request) {
	http.Error(request.Response(), "Not Found", 404)
}

func render498(request *prago.Request) {
	http.Error(request.Response(), "Wrong Hash", 498)
}
