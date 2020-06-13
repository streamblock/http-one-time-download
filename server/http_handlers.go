package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)


func sendJSONResponseSimpleText(w http.ResponseWriter, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)

	if err := json.NewEncoder(w).Encode(jsonMessageSimple{Code: httpStatus, Text: message}); err != nil {
		panic(err)
	}
}

func sendJSONResponseUploadSuccess(w http.ResponseWriter, httpStatus int, jsonData jsonMessageUploadSuccess) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)

	if err := json.NewEncoder(w).Encode(jsonData); err != nil {
		panic(err)
	}
}

func sendHTMLResponse(w http.ResponseWriter, message string, httpStatus int) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(httpStatus)

	w.Write([]byte(message))
}

func handleRegularUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

		t, err := template.ParseFiles("www/upload.html")
		if err != nil {
			sendHTMLResponse(w, "Unable to load template", http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	}

	if r.Method == "POST" {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			sendJSONResponseSimpleText(w, "Could not parse multipart form", http.StatusInternalServerError)
			log.Println("Could not parse multipart form: ", err)
			return
		}
		// Parse and validate file and post parameters
		file, fileHeader, err := r.FormFile("uploadFile")
		if err != nil {
			sendJSONResponseSimpleText(w, "Form error", http.StatusUnprocessableEntity)
			log.Println("Form error: ", err)
			return
		}
		defer file.Close()

		if filePath, sha256sum, err := storeUploadFile(file, fileHeader); err != nil {
			sendJSONResponseSimpleText(w, "Upload failed (generic error)", http.StatusInternalServerError)
		} else {
			sendJSONResponseUploadSuccess(w, http.StatusCreated, jsonMessageUploadSuccess{Code: http.StatusCreated, UploadPath: filePath, FileChecksum: sha256sum})
		}
	}
}

func handleSecureUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "www/upload-secu.html")
	}
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	// Cleanup requested URL path
	log.Println(r.URL.Path)
	cleanURLPath := path.Clean(r.URL.Path)
	cleanURLPath = strings.Replace(cleanURLPath, downloadURLPattern, "", 1)

	// Check if file exist
	localFilePath := filepath.Join(uploadPath, cleanURLPath)
	if _, err := os.Stat(localFilePath); err == nil {
		log.Println("File exists: ", localFilePath)
		// Serve the file to client
		http.ServeFile(w, r, localFilePath)
		log.Println("File served: ", localFilePath)
		// Delete it from file system
		localDirPath := filepath.Dir(localFilePath)
		os.RemoveAll(localDirPath)
		log.Println("Directory deleted: ", localDirPath)
		return
	} else if os.IsNotExist(err) {
		log.Println("File does not exist: ", localFilePath)
		sendJSONResponseSimpleText(w, "Internal Error", http.StatusInternalServerError)
		return
	} else {
		log.Println("File may or may not exist: ", err)
		sendJSONResponseSimpleText(w, "Internal Error", http.StatusInternalServerError)
		return
	}
}