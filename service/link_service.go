package service

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"github.com/vldmir/zip-service/config"
)

type Task struct {
	ID    string
	Links []string
}

type LinkService struct {
	tasks map[string]*Task
	mu    sync.RWMutex
	cfg   *config.Config
}

func New(cfg *config.Config) *LinkService {
	return &LinkService{
		tasks: make(map[string]*Task),
		cfg:   cfg,
	}
}

// CreateTask создает новую задачу и возвращает её UUID
func (ls *LinkService) CreateTask() string {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	taskID := uuid.New().String()
	ls.tasks[taskID] = &Task{
		ID:    taskID,
		Links: make([]string, 0),
	}
	return taskID
}

// AddLink добавляет ссылку в указанную задачу
func (ls *LinkService) AddLink(taskID, link string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	task, exists := ls.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	task.Links = append(task.Links, link)
	return nil
}

// GetLinks возвращает ссылки для указанной задачи
func (ls *LinkService) GetLinks(taskID string) ([]string, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	task, exists := ls.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	return task.Links, nil
}

// ClearTask очищает указанную задачу
func (ls *LinkService) ClearTask(taskID string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if _, exists := ls.tasks[taskID]; !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	delete(ls.tasks, taskID)
	return nil
}

// GetTaskStatus возвращает информацию о задаче
func (ls *LinkService) GetTaskStatus(taskID string) (int, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	task, exists := ls.tasks[taskID]
	if !exists {
		return 0, fmt.Errorf("task with ID %s not found", taskID)
	}

	return len(task.Links), nil
}

func (ls *LinkService) ActiveTasksCount() int {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return len(ls.tasks)
}