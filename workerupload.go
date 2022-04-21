package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func computeContentMD5(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, nil
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func computeContentBase64MD5(name string) (string, error) {
	md5, err := computeContentMD5(name)
	if err != nil {
		return "", nil
	}
	return base64.StdEncoding.EncodeToString(md5), nil
}

type FileUploader interface {
	UploadFile(src, dst string, meta map[string]string) error
}

type MockUploader struct {
	Error error
}

func (up *MockUploader) UploadFile(src, dst string, meta map[string]string) error {
	if up.Error != nil {
		return up.Error
	}
	log.Printf("mock uploading file:\nsrc: %s\ndst: %s\nmeta: %+v", src, dst, meta)
	// TODO(sean) track results in a list or map to be used in tests
	return nil
}

type S3Uploader struct{}

func (up *S3Uploader) UploadFile(src, dst string, meta map[string]string) error {
	// TODO(sean) is this the right interface? maybe we can more closely match the s3 UploadInput?
	log.Printf("uploading file to s3: %s -> %s", src, dst)

	contentMD5, err := computeContentBase64MD5(src)
	if err != nil {
		return err
	}

	// optional - check if file already exists or if content hash matches
	// if uploadExistsInS3(filepath.Join(s3path, targetNameData)) {
	// 	fmt.Println("Files already exist in S3")
	// 	return nil
	// }

	uploader := s3manager.NewUploader(newSession)

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3manager/#Uploader
	upi := &s3manager.UploadInput{
		Bucket:     aws.String(s3bucket),
		Key:        aws.String(dst),
		Body:       f,
		ContentMD5: aws.String(contentMD5),
		Metadata:   aws.StringMap(meta),
	}

	// Upload the file to S3.
	result, err := uploader.Upload(upi)
	if err != nil {
		return err
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)

	// TODO(sean) confirm file or content length can be read back out?
	// TODO(sean) the return string was left as missing before. why is this?
	return nil
}

// prefix is simply the path:  <path>/<targetFilename>
func uploadFileToS3(prefix string, filename string, targetFilename string, meta map[string]string) (string, error) {
	dst := filepath.Join(prefix, targetFilename)

	up := &MockUploader{}
	// up := &MockUploader{Error: fmt.Errorf("there was a problem uploading the file")}

	// up := &S3Uploader{}
	return "", up.UploadFile(filename, dst, meta)
}
