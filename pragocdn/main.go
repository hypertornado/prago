package main

import (
	"errors"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/pragocdn/cdnclient"
	"github.com/hypertornado/prago/utils"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const version = "0.1.0"

var config CDNConfig

var accounts = map[string]*CDNConfigAccount{}
var homePath = os.Getenv("HOME")

var uuidRegex = regexp.MustCompile("^[a-zA-Z0-9]{10,}$")
var filenameRegex = regexp.MustCompile("^[a-zA-Z0-9_-]{1,50}$")
var extensionRegex = regexp.MustCompile("^[a-zA-Z0-9]{1,10}$")

func main() {
	var err error
	config, err = loadCDNConfig()
	if err != nil {
		panic(err)
	}

	for _, v := range config.Accounts {
		err := prepareAccountDirectories(v.Name)
		if err != nil {
			panic(err)
		}
		accounts[v.Name] = &v
	}

	app := prago.NewApp("pragocdn", version)
	app.AddMiddleware(prago.MiddlewareServer{Fn: start})
	prago.Must(app.Init())
}

func uploadFile(account CDNConfigAccount, extension string, inData io.Reader) (*cdnclient.CDNUploadData, error) {
	uuid := utils.RandomString(20)
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

	_, err = io.Copy(file, inData)
	if err != nil {
		return nil, err
	}

	return &cdnclient.CDNUploadData{
		UUID:      uuid,
		Extension: extension,
	}, nil
}

func start(app *prago.App) {

	app.MainController().Post("/:account/upload/:extension", func(request prago.Request) {
		defer request.Request().Body.Close()
		accountName := request.Params().Get("account")
		account := accounts[accountName]
		if account == nil {
			panic("no account")
		}

		authorization := request.Request().Header.Get("X-Authorization")
		if account.Password != authorization {
			panic("wrong authorization")
		}

		extension := normalizeExtension(request.Params().Get("extension"))
		if !extensionRegex.MatchString(extension) {
			panic("wrong extension")
		}

		data, err := uploadFile(*account, extension, request.Request().Body)
		if err != nil {
			panic(err)
		}
		prago.WriteAPI(request, data, 200)
	})

	app.MainController().Get("/:account/:uuid/:format/:name", func(request prago.Request) {
		errCode, err, stream := getFile(
			request.Params().Get("account"),
			request.Params().Get("uuid"),
			request.Params().Get("format"),
			request.Params().Get("name"),
		)

		if err != nil {
			panic(err)
		}

		switch errCode {
		case 404:
			render404(request)
			return
		}

		_, err = io.Copy(request.Response(), stream)
		if err != nil {
			panic(err)
		}

		return
	})
}

var fileExtensionMap = map[string]string{
	"jpeg": "jpg",
}

func getFile(accountName, uuid, format, name string) (eddCode int, err error, source io.Reader) {
	account := accounts[accountName]
	if account == nil {
		return 404, errors.New("account not found"), nil
	}

	if !uuidRegex.MatchString(uuid) {
		return 404, errors.New("wrongs uuid format: " + uuid), nil
	}

	splited := strings.Split(name, ".")
	if len(splited) != 2 {
		return 404, errors.New("wrong name format"), nil
	}
	fileName := splited[0]
	fileExtension := splited[1]
	fileExtension = normalizeExtension(fileExtension)

	if !filenameRegex.MatchString(fileName) || !extensionRegex.MatchString(fileExtension) {
		return 404, errors.New("wrong name format"), nil
	}

	originalPath := getFilePath(accountName, uuid, fileExtension)

	var path string
	if format == "file" {
		path = originalPath
	} else {
		path, err = convertedFilePath(accountName, uuid, fileExtension, format)
		if err != nil {
			return 404, err, nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return 500, err, nil
	}

	return 200, nil, file
}

func prepareAccountDirectories(name string) error {
	var err error
	err = os.MkdirAll(
		fmt.Sprintf("%s/.pragocdn/files/%s",
			homePath,
			name,
		),
		0777,
	)
	if err != nil {
		return fmt.Errorf("preparing files dir for %s: %s", name, err)
	}

	err = os.MkdirAll(
		fmt.Sprintf("%s/.pragocdn/cache/%s",
			homePath,
			name,
		),
		0777,
	)
	if err != nil {
		return fmt.Errorf("preparing cache dir for %s: %s", name, err)
	}
	return nil
}

func getFileDirectoryPath(account, uuid string) string {
	firstPrefix := uuid[0:2]
	secondPrefix := uuid[2:4]
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

func getCacheFilePath(account, uuid, format, extension string) string {
	return fmt.Sprintf("%s/.pragocdn/cache/%s/%s_%s.%s",
		homePath,
		account,
		uuid,
		format,
		extension,
	)
}

var singleSizeRegexp = regexp.MustCompile("^[1-9][0-9]*$")
var sizeRegexp = regexp.MustCompile("^[1-9][0-9]{0,3}x[1-9][0-9]{0,3}$")

func isImageExtension(extension string) bool {
	if extension == "jpg" || extension == "png" {
		return true
	}
	return false
}

func normalizeExtension(extension string) string {
	extension = strings.ToLower(extension)
	fileExtensionChanged := fileExtensionMap[extension]
	if fileExtensionChanged != "" {
		extension = fileExtensionChanged
	}
	return extension
}

func convertedFilePath(account, uuid, extension, format string) (string, error) {
	if !isImageExtension(extension) {
		return "", errors.New("cant resize non images")
	}

	originalPath := getFilePath(account, uuid, extension)
	outputPath := getCacheFilePath(account, uuid, format, extension)

	if singleSizeRegexp.MatchString(format) {
		return outputPath, vipsThumbnail(originalPath, outputPath, format, false)
	}

	if sizeRegexp.MatchString(format) {
		return outputPath, vipsThumbnail(originalPath, outputPath, format, true)
	}

	return "", errors.New("wrong file convert format")
}

func vipsThumbnail(originalPath, outputPath, size string, crop bool) error {
	_, err := os.Open(outputPath)
	if err == nil {
		return nil
	}

	cmdAr := []string{
		originalPath,
		"-s",
		size,
		"-o",
		outputPath + "[optimize_coding,strip]",
	}

	if config.Profile != "" {
		cmdAr = append(cmdAr, "--delete", "--eprofile", config.Profile)
	}

	if crop {
		cmdAr = append(cmdAr, "-m", "attention")
	}

	cmd := exec.Command("vipsthumbnail", cmdAr...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func render404(request prago.Request) {
	request.SetProcessed()
	http.NotFound(request.Response(), request.Request())
}
