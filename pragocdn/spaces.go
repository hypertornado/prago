package main

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (project *CDNProject) hasSpacesConfig() bool {
	return project.CDNEndpointURL != "" &&
		project.CDNAccessKey != "" &&
		project.CDNSecretKey != ""
	//project.CDNRegion != "" &&
	//project.CDNBucket != ""
}

func (project *CDNProject) newSpacesClient() *s3.Client {
	return s3.New(s3.Options{
		BaseEndpoint: aws.String(project.CDNEndpointURL),
		Region:       project.CDNRegion,
		UsePathStyle: true,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				project.CDNAccessKey,
				project.CDNSecretKey,
				"",
			),
		),
	})
}

func (project *CDNProject) deleteFolderFromSpaces(pathPrefix string) error {
	return fmt.Errorf("deleting from spaces cdn not implemented yet")
}

// Copy files form localfolder to spaces using uploadFileToSpaces
func (project *CDNProject) uploadFolderToSpaces(localFolderPath string, pathPrefix string) error {
	return filepath.Walk(localFolderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(localFolderPath, filePath)
		if err != nil {
			return err
		}
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		if pathPrefix != "" {
			rel = pathPrefix + rel
		}

		fmt.Println("REL", rel)

		uploadErr := project.uploadFileToSpaces(rel, contentType, f, info.Size())
		f.Close()
		return uploadErr
	})
}

func (project *CDNProject) uploadFileToSpaces(key, contentType string, r io.Reader, size int64) error {
	if !project.hasSpacesConfig() {
		return nil
	}

	fmt.Println("UPLOADING FILE", project.Name, key)

	client := project.newSpacesClient()
	res, err := client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(project.Name),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
		ACL:           "public-read",
	})
	if err != nil {
		return fmt.Errorf("spaces: put file: %w", err)
	}

	_ = res
	fmt.Printf("FILE URL: %s/%s/%s\n", project.CDNEndpointURL, project.Name, key)

	return nil
}

func (project *CDNProject) uploadVideoToSpaces(key, filename string, r io.Reader, size int64) error {
	contentType := mime.TypeByExtension(path.Ext(filename))
	if contentType == "" {
		contentType = "video/mp4"
	}
	return project.uploadFileToSpaces(key, contentType, r, size)
}
