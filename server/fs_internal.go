package main

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
)


func randToken(len int64) string {
	b := make([]byte, len)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func storeUploadFile(file multipart.File, fileHeader *multipart.FileHeader) (filePath string, sha256sum string, err error) {
	// Validate file size
	fileSize := fileHeader.Size
	if fileSize > maxUploadSize {
		log.Printf("Upload file size too big: got %v bytes, max %v bytes\n", fileSize, maxUploadSize)
		return "", "", errors.New("File too big")
	}
	log.Println(fileSize, maxUploadSize)

	// Generate random subdir for the uploaded file
	subdirName := randToken(uploadSubDirLength)
	uploadDirPath := filepath.Join(uploadPath, subdirName)
	if err := ensureDir(uploadDirPath); err != nil {
		log.Println("Directory creation failed with error: " + err.Error())
		return "", "", err
	}

	// Write new file
	uploadFileName := filepath.Join(uploadDirPath, path.Clean(fileHeader.Filename))
	uploadFile, err := os.Create(uploadFileName)
	if err != nil {
		os.Remove(uploadDirPath)
		log.Println("Dest File creation failed with error: " + err.Error())
		return "", "", err
	}
	defer uploadFile.Close()
	_, err = io.Copy(uploadFile, file)
	if err != nil {
		os.Remove(uploadDirPath)
		log.Println("Cannot write file: " + err.Error())
		return "", "", err
	}
	uploadFile.Sync()
	uploadFile.Seek(0, io.SeekStart)
	h := sha256.New()
	if _, err = io.Copy(h, uploadFile); err != nil {
		return "", "", err
	}
	sha256sum = fmt.Sprintf("%x", h.Sum(nil))
	log.Println("File successfully created: ", uploadFileName)
	return strings.Replace(uploadFileName, uploadPath, "", 1), sha256sum, nil
}
