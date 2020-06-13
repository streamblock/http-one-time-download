package main

type jsonMessageSimple struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type jsonMessageUploadSuccess struct {
	Code 			int 	`json:"code"`
	UploadPath 		string 	`json:"path"`
	FileChecksum	string 	`json:"sha256sum"`
}