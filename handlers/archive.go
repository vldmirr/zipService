package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/vldmir/zip-service/manager"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"github.com/vldmir/zip-service/config"
	"github.com/vldmir/zip-service/service"
)

var (
	storage *service.LinkService
	cfg     *config.Config
)

func InitHandlers(config *config.Config) {
	cfg = config
	storage = service.New(config) // Инициализируем storage с конфигом
}

type TaskResponse struct {
	TaskID string `json:"task_id"`
}

// CreateTaskHandler создает новую задачу для загрузки файлов
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := storage.CreateTask()
	json.NewEncoder(w).Encode(TaskResponse{TaskID: taskID})
}

// AddLinkHandler добавляет ссылку в указанную задачу
func AddLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	var data struct {
		Link string `json:"link"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := storage.AddLink(taskID, data.Link); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DownloadAndArchiveHandler обрабатывает загрузку и архивацию файлов для задачи
func DownloadAndArchiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Получаем ссылки для задачи
	linkStrings, err := storage.GetLinks(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Конвертируем строки в URL
	var links []*url.URL
	for _, linkStr := range linkStrings {
		u, err := url.Parse(linkStr)
		if err != nil {
			log.Printf("Error parsing URL %s: %v", linkStr, err)
			continue
		}
		links = append(links, u)
	}


	// Указываем директорию для загрузки
	downloadDir := "./downloads"
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		if err := os.Mkdir(downloadDir, 0755); err != nil {
			log.Printf("Failed to create download directory: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Загружаем файлы
	manager.Run(links, downloadDir)

	// Создаем архив
	archiveName := "downloads.zip"
	if filename := r.URL.Query().Get("filename"); filename != "" {
		archiveName = filename
		if !strings.HasSuffix(archiveName, ".zip") {
			archiveName += ".zip"
		}
	}

	// Устанавливаем заголовки для потоковой передачи архива
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", archiveName))

	// Создаем ZIP-архив напрямую в ResponseWriter
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Проходим по файлам в папке
	err = filepath.Walk(downloadDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем директории и архивы
		if info.IsDir() || strings.HasSuffix(filePath, ".zip") {
			return nil
		}

		fileToZip, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", filePath, err)
		}
		defer fileToZip.Close()

		// Создаем запись в архиве
		relPath, _ := filepath.Rel(downloadDir, filePath)
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %s: %v", filePath, err)
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create writer for %s: %v", filePath, err)
		}

		_, err = io.Copy(writer, fileToZip)
		return err
	})

	if err != nil {
		log.Printf("Error during archiving: %v", err)
		// Мы не можем изменить статус ответа, так как уже начали передачу данных
		fmt.Fprintf(w, "\nError during archiving: %v", err)
	}
}

// GetTaskStatusHandler возвращает статус задачи
func GetTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	count, err := storage.GetTaskStatus(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"links_count": count})
}
