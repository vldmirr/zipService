package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vldmir/zip-service/handlers"
)

func printWelcomeMessage() {
	fmt.Println(`
            ##

 ######    ###     ######             #####    ####    ######   ##  ##    ####    ######
 #  ##      ##      ##  ##  ######   ##       ##  ##    ##  ##  ##  ##   ##  ##    ##  ##
   ##       ##      ##  ##            #####   ######    ##      ##  ##   ######    ##
  ##  #     ##      #####                 ##  ##        ##       ####    ##        ##
 ######    ####     ##               ######    #####   ####       ##      #####   ####
                   ####

	`)
	fmt.Println("ZIP File Service v1.0")
	fmt.Printf("Server starting at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("----------------------------------------")
}

func printRoutes() {
	fmt.Println("Available endpoints:")
	fmt.Println("POST   /task/create              - Create new download task")
	fmt.Println("POST   /links/add?task=<task_id> - Add link to task")
	fmt.Println("GET    /task/status?task=<task_id> - Check task status")
	fmt.Println("GET    /task/download-archive?task=<task_id> - Download archive")
	fmt.Println("----------------------------------------")
}

func printServerInfo() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	fmt.Printf("Server running on:\n")
	fmt.Printf("• Local: http://localhost:8080\n")
	fmt.Printf("• Network: http://%s:8080\n", hostname)
	fmt.Println("----------------------------------------")
	fmt.Println("Press Ctrl+C to stop the server")
}

func main() {
	printWelcomeMessage()
	printRoutes()
	printServerInfo()

	http.HandleFunc("/task/create", handlers.CreateTaskHandler)
	http.HandleFunc("/links/add", handlers.AddLinkHandler)
	http.HandleFunc("/task/download-archive", handlers.DownloadAndArchiveHandler)
	http.HandleFunc("/task/status", handlers.GetTaskStatusHandler)

	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}