package handlers

import (
	"encoding/json"
	"github.com/vldmir/zip-service/service"
	"net/http"
)

type LinkServiceHandler struct {
	service *service.LinkService
}

func NewLinkServiceHandler(s *service.LinkService) *LinkServiceHandler {
	return &LinkServiceHandler{service: s}
}

// CreateTaskHandler создает новую задачу
func (h *LinkServiceHandler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := h.service.CreateTask()
	json.NewEncoder(w).Encode(map[string]string{"task_id": taskID})
}

// AddLinkHandler добавляет ссылку в задачу
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

	link := r.URL.Query().Get("link")
	if link == "" {
		http.Error(w, "Link is required", http.StatusBadRequest)
		return
	}

	if err := h.service.AddLink(taskID, link); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
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
