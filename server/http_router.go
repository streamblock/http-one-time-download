package main

import (
	"net/http"
)

func newMux(config *configuration) *http.ServeMux {
	var handler http.Handler

	mux := http.NewServeMux()
	// static/js static/css content
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("www/static"))))

	// download
	handler = http.HandlerFunc(handleDownload)
	if (config.Logging.LogEveryRequest) {
		handler = httpLogger(handler, "Download")
	}
	mux.Handle(config.Server.Download.URL, handler)

	// regular upload
	if (config.Server.RegularUpload.Enable) {
		var handler http.Handler
		handler = http.HandlerFunc(handleRegularUploadFile)
		if (config.Logging.LogEveryRequest) {
			handler = httpLogger(handler, "RegularUpload")
		}
		mux.Handle(config.Server.RegularUpload.URL, handler)
	}

	// secure upload
	if (config.Server.SecureUpload.Enable) {
		var handler http.Handler
		handler = http.HandlerFunc(handleSecureUploadFile)
		if (config.Logging.LogEveryRequest) {
			handler = httpLogger(handler, "SecureUpload")
		}
		mux.Handle(config.Server.SecureUpload.URL, handler)
	}

	return mux
}
