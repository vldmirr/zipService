package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/vldmir/zip-service/config"
	"github.com/vldmir/zip-service/service"
)

type LinkServiceHandler struct {
	service *service.LinkService
	config  *config.Config
}

func NewLinkServiceHandler(s *service.LinkService,cfg *config.Config) *LinkServiceHandler {
	return &LinkServiceHandler{
		service: s,
		config:  cfg,
	}
}


func (h *LinkServiceHandler) AddLinkHandler(w http.ResponseWriter, r *http.Request) {
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

	// Проверка типа файла
	fileType := strings.ToLower(filepath.Ext(data.Link))
	allowed := false
	for _, t := range h.config.AllowedTypes {
		if fileType == t {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "File type not allowed", http.StatusBadRequest)
		return
	}

	// Проверка лимита файлов в задаче
	if count, err := h.service.GetTaskStatus(taskID); err == nil {
		if count >= h.config.Limits.MaxFilesPerTask {
			http.Error(w, "Task file limit reached", http.StatusBadRequest)
			return
		}
	}

	if err := h.service.AddLink(taskID, data.Link); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetLinksHandler возвращает ссылки задачи
func (h *LinkServiceHandler) GetLinksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	links, err := h.service.GetLinks(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(links)
}

// TaskStatusHandler возвращает статус задачи
func (h *LinkServiceHandler) TaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	count, err := h.service.GetTaskStatus(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"links_count": count})
}
