package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/middleware"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/response"
	"github.com/taskflow/backend/internal/service"
)

type ProjectHandler struct {
	projectService *service.ProjectService
	taskService    *service.TaskService
}

func NewProjectHandler(projectService *service.ProjectService, taskService *service.TaskService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService, taskService: taskService}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := make(map[string]string)
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "is required"
	}
	if len(fields) > 0 {
		response.ValidationError(w, fields)
		return
	}

	userID := middleware.GetUserID(r.Context())
	project, err := h.projectService.Create(r.Context(), req, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	params := parsePagination(r)

	projects, total, err := h.projectService.List(r.Context(), userID, params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	pagination := model.NewPagination(params, total)
	response.JSONList(w, http.StatusOK, projects, pagination)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id format")
		return
	}

	userID := middleware.GetUserID(r.Context())
	project, err := h.projectService.Get(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Fetch tasks for this project (no filters, no pagination — return all)
	tasks, _, err := h.taskService.List(r.Context(), id, nil, nil, model.PaginationParams{Page: 1, Limit: 100}, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, model.NewProjectDetail(project, tasks))
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id format")
		return
	}

	var req model.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := middleware.GetUserID(r.Context())
	project, err := h.projectService.Update(r.Context(), id, req, userID)
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

	response.JSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id format")
		return
	}

	userID := middleware.GetUserID(r.Context())
	err = h.projectService.Delete(r.Context(), id, userID)
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

func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id format")
		return
	}

	userID := middleware.GetUserID(r.Context())
	stats, err := h.taskService.Stats(r.Context(), id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, stats)
}

func parsePagination(r *http.Request) model.PaginationParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	return model.NewPaginationParams(page, limit)
}
