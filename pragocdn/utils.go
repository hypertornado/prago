package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func getNameAndExtension(filename string) (name, extension string, err error) {
	extension = filepath.Ext(filename)
	if extension == "" {
		return "", "", errors.New("no extension")
	}
	extension = extension[1:]

	name = filename[0 : len(filename)-len(extension)-1]

	if !filenameRegex.MatchString(name) {
		return "", "", errors.New("wrong name of file")
	}

	if !extensionRegex.MatchString(extension) {
		return "", "", errors.New("wrong extension of file")
	}

	extension = normalizeExtension(extension)

	return name, extension, nil
}

func unlocalized(in string) func(string) string {
	return func(string) string {
		return in
	}
}

func normalizeExtension(extension string) string {
	extension = strings.ToLower(extension)
	fileExtensionChanged := fileExtensionMap[extension]
	if fileExtensionChanged != "" {
		extension = fileExtensionChanged
	}
	return extension
}

func cdnDirPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDir + "/.pragocdn"
}
