package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
)

//https://github.com/muesli/smartcrop

// CMYK: https://github.com/jcupitt/libvips/issues/630
func vipsThumbnail(originalPath, outputFilePath, size string, crop bool) error {
	n := rand.Int() % len(vipsMutexes)
	vipsMutex := vipsMutexes[n]
	vipsMutex.Lock()
	defer vipsMutex.Unlock()

	extension := "webp"

	f, err := os.Open(outputFilePath)
	if err == nil {
		f.Close()
		return nil
	}

	outputDirectoryPath := getDirFromFilepath(outputFilePath)

	err = os.MkdirAll(outputDirectoryPath, 0777)
	if err != nil {
		return fmt.Errorf("error while creating mkdirall %s: %s", outputDirectoryPath, err)
	}

	tempPath := getTempFilePath(extension)
	defer os.Remove(tempPath)

	err = vipsThumbnailProfile(originalPath, tempPath, size, crop, false)
	if err != nil {
		err = vipsThumbnailProfile(originalPath, tempPath, size, crop, true)
	}
	if err != nil {
		return fmt.Errorf("vipsThumbnailProfile: %s", err)
	}

	err = os.Rename(tempPath, outputFilePath)
	if err != nil {
		return fmt.Errorf("moving file from %s to %s: %s", tempPath, outputFilePath, err)
	}

	return nil
}

func vipsThumbnailProfile(originalPath, outputFilePath, size string, crop bool, cmyk bool) error {

	cmdAr := []string{
		originalPath,
		"--rotate",
		"-s",
		size,
		"-o",
		outputFilePath + "[strip]",
	}

	if cmyk {
		cmdAr = append(cmdAr, "-i", cdnDirPath()+"/cmyk.icm")
	}

	if crop {
		cmdAr = append(cmdAr, "--smartcrop", "attention")
	}

	var b bytes.Buffer

	cmd := exec.Command("vipsthumbnail", cmdAr...)
	cmd.Stdout = &b
	cmd.Stderr = &b

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("vips exited with error: %s, output: %s;", err, string(b.Bytes()))
	}
	return nil
}

func getTempFilePath(extension string) string {
	dir := os.TempDir()
	fileName := fmt.Sprintf("pragocdn-%d.%s", rand.Int(), extension)
	return path.Join(dir, fileName)
}
