package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/taskflow/backend/internal/model"
)

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *model.Project) error {
	query := `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, project.Name, project.Description, project.OwnerID).
		Scan(&project.ID, &project.CreatedAt)
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var project model.Project
	query := `SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`
	err := r.db.GetContext(ctx, &project, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &project, err
}

func (r *ProjectRepository) List(ctx context.Context, userID uuid.UUID, params model.PaginationParams) ([]model.Project, int, error) {
	countQuery := `
		SELECT COUNT(DISTINCT p.id) FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, userID); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`
	var projects []model.Project
	if err := r.db.SelectContext(ctx, &projects, query, userID, params.Limit, params.Offset()); err != nil {
		return nil, 0, err
	}
	if projects == nil {
		projects = []model.Project{}
	}
	return projects, total, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *model.Project) error {
	query := `
		UPDATE projects SET name = $1, description = $2
		WHERE id = $3
		RETURNING created_at`
	result, err := r.db.ExecContext(ctx, query, project.Name, project.Description, project.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProjectRepository) UserHasAccessToProject(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM projects WHERE id = $1 AND owner_id = $2
			UNION
			SELECT 1 FROM tasks WHERE project_id = $1 AND assignee_id = $2
		)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, projectID, userID)
	return exists, err
}
