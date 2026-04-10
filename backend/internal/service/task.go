package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

type TaskService struct {
	taskRepo    *repository.TaskRepository
	projectRepo *repository.ProjectRepository
}

func NewTaskService(taskRepo *repository.TaskRepository, projectRepo *repository.ProjectRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo, projectRepo: projectRepo}
}

func (s *TaskService) Create(ctx context.Context, projectID uuid.UUID, req model.CreateTaskRequest, userID uuid.UUID) (*model.Task, error) {
	// Verify project exists
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		return nil, err
	}

	// Verify user has access
	hasAccess, err := s.projectRepo.UserHasAccessToProject(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	// For task creation, also allow the project owner (covered by UserHasAccessToProject)
	// and any authenticated user who can see the project
	project, _ := s.projectRepo.GetByID(ctx, projectID)
	if !hasAccess && project.OwnerID != userID {
		return nil, ErrForbidden
	}

	priority := req.Priority
	if priority == "" {
		priority = model.TaskPriorityMedium
	}

	task := &model.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatusTodo,
		Priority:    priority,
		ProjectID:   projectID,
		AssigneeID:  req.AssigneeID,
		CreatedBy:   userID,
	}

	if req.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			task.DueDate = sql.NullTime{Time: parsed, Valid: true}
		}
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) List(ctx context.Context, projectID uuid.UUID, status *model.TaskStatus, assigneeID *uuid.UUID, params model.PaginationParams, userID uuid.UUID) ([]model.Task, int, error) {
	// Verify access
	hasAccess, err := s.projectRepo.UserHasAccessToProject(ctx, projectID, userID)
	if err != nil {
		return nil, 0, err
	}
	project, pErr := s.projectRepo.GetByID(ctx, projectID)
	if pErr != nil {
		return nil, 0, pErr
	}
	if !hasAccess && project.OwnerID != userID {
		return nil, 0, repository.ErrNotFound
	}

	return s.taskRepo.List(ctx, projectID, status, assigneeID, params)
}

func (s *TaskService) Update(ctx context.Context, taskID uuid.UUID, req model.UpdateTaskRequest, userID uuid.UUID) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Verify user has access to the project
	hasAccess, err := s.projectRepo.UserHasAccessToProject(ctx, task.ProjectID, userID)
	if err != nil {
		return nil, err
	}
	project, _ := s.projectRepo.GetByID(ctx, task.ProjectID)
	if !hasAccess && project.OwnerID != userID {
		return nil, ErrForbidden
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			task.DueDate = sql.NullTime{Time: parsed, Valid: true}
		}
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Only project owner or task creator can delete
	project, err := s.projectRepo.GetByID(ctx, task.ProjectID)
	if err != nil {
		return err
	}

	if project.OwnerID != userID && task.CreatedBy != userID {
		return ErrForbidden
	}

	return s.taskRepo.Delete(ctx, taskID)
}

func (s *TaskService) Stats(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (*model.ProjectStats, error) {
	// Verify access
	hasAccess, err := s.projectRepo.UserHasAccessToProject(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	project, pErr := s.projectRepo.GetByID(ctx, projectID)
	if pErr != nil {
		return nil, pErr
	}
	if !hasAccess && project.OwnerID != userID {
		return nil, repository.ErrNotFound
	}

	return s.taskRepo.GetStatsByProject(ctx, projectID)
}
