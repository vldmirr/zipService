package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/vldmir/zip-service/config"
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

func printServerInfo(port string) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	fmt.Printf("Server running on:\n")
	fmt.Printf("• Local: http://localhost:%s\n",port)
	fmt.Printf("• Network: http://%s:\n", hostname)
	fmt.Println("----------------------------------------")
	fmt.Println("Press Ctrl+C to stop the server")
}

func main() {
		// Загрузка конфигурации
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация обработчиков с конфигом
	handlers.InitHandlers(cfg)

	// Настройка сервера с таймаутами
	srv := &http.Server{
		Addr:         cfg.Server.Port,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	printWelcomeMessage()
	printRoutes()
	printServerInfo(cfg.Server.Port)

	http.HandleFunc("/task/create", handlers.CreateTaskHandler)
	http.HandleFunc("/links/add", handlers.AddLinkHandler)
	http.HandleFunc("/task/download-archive", handlers.DownloadAndArchiveHandler)
	http.HandleFunc("/task/status", handlers.GetTaskStatusHandler)

	log.Printf("Starting server on %s with configuration:\n", cfg.Server.Port)
	log.Printf("- Max concurrent tasks: %d\n", cfg.Limits.MaxConcurrentTasks)
	log.Printf("- Max files per task: %d\n", cfg.Limits.MaxFilesPerTask)
	log.Printf("- Allowed file types: %v\n", cfg.AllowedTypes)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}