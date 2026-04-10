package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/repository"
)

var ErrForbidden = errors.New("forbidden")

type ProjectService struct {
	projectRepo *repository.ProjectRepository
}

func NewProjectService(projectRepo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

func (s *ProjectService) Create(ctx context.Context, req model.CreateProjectRequest, ownerID uuid.UUID) (*model.Project, error) {
	project := &model.Project{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	}
	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.projectRepo.UserHasAccessToProject(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, repository.ErrNotFound
	}

	return project, nil
}

func (s *ProjectService) List(ctx context.Context, userID uuid.UUID, params model.PaginationParams) ([]model.Project, int, error) {
	return s.projectRepo.List(ctx, userID, params)
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, req model.UpdateProjectRequest, userID uuid.UUID) (*model.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project.OwnerID != userID {
		return nil, ErrForbidden
	}

	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = req.Description
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if project.OwnerID != userID {
		return ErrForbidden
	}
	return s.projectRepo.Delete(ctx, id)
}
