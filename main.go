package main

import (
	"github.com/vldmir/zip-service/handlers"

	"net/http"
)

func main() {
	http.HandleFunc("/task/create", handlers.CreateTaskHandler)
	http.HandleFunc("/links/add", handlers.AddLinkHandler)
	http.HandleFunc("/task/download-archive", handlers.DownloadAndArchiveHandler)
	http.HandleFunc("/task/status", handlers.GetTaskStatusHandler)

	http.ListenAndServe(":8080", nil)
}
