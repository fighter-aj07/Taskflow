package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/middleware"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/response"
	"github.com/taskflow/backend/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id format")
		return
	}

	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	fields := make(map[string]string)
	if req.Title == "" {
		fields["title"] = "is required"
	}
	if req.Priority != "" && !req.Priority.IsValid() {
		fields["priority"] = "must be low, medium, or high"
	}
	if len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	userID := middleware.GetUserID(r.Context())
	task, err := h.taskService.Create(r.Context(), projectID, req, userID)
	if errors.Is(err, repository.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "not found")
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id format")
		return
	}

	// Parse optional filters
	var status *model.TaskStatus
	if s := r.URL.Query().Get("status"); s != "" {
		ts := model.TaskStatus(s)
		if !ts.IsValid() {
			response.Error(w, http.StatusBadRequest, "invalid status filter")
			return
		}
		status = &ts
	}

	var assigneeID *uuid.UUID
	if a := r.URL.Query().Get("assignee"); a != "" {
		parsed, err := uuid.Parse(a)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid assignee filter")
			return
		}
		assigneeID = &parsed
	}

	userID := middleware.GetUserID(r.Context())
	params := parsePagination(r)

	tasks, total, err := h.taskService.List(r.Context(), projectID, status, assigneeID, params, userID)
	if errors.Is(err, repository.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	pagination := model.NewPagination(params, total)
	response.JSONList(w, http.StatusOK, tasks, pagination)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id format")
		return
	}

	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate enums if provided
	fields := make(map[string]string)
	if req.Status != nil && !req.Status.IsValid() {
		fields["status"] = "must be todo, in_progress, or done"
	}
	if req.Priority != nil && !req.Priority.IsValid() {
		fields["priority"] = "must be low, medium, or high"
	}
	if len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	userID := middleware.GetUserID(r.Context())
	task, err := h.taskService.Update(r.Context(), taskID, req, userID)
	if errors.Is(err, repository.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "not found")
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id format")
		return
	}

	userID := middleware.GetUserID(r.Context())
	err = h.taskService.Delete(r.Context(), taskID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "not found")
		return
	}
	if errors.Is(err, service.ErrForbidden) {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.NoContent(w)
}
