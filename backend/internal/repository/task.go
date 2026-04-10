package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/taskflow/backend/internal/model"
)

type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, created_by, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		task.Title, task.Description, task.Status, task.Priority,
		task.ProjectID, task.AssigneeID, task.CreatedBy, task.DueDate,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var task model.Task
	query := `SELECT id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at FROM tasks WHERE id = $1`
	err := r.db.GetContext(ctx, &task, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &task, err
}

func (r *TaskRepository) List(ctx context.Context, projectID uuid.UUID, status *model.TaskStatus, assigneeID *uuid.UUID, params model.PaginationParams) ([]model.Task, int, error) {
	where := []string{"project_id = $1"}
	args := []interface{}{projectID}
	argIdx := 2

	if status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *status)
		argIdx++
	}
	if assigneeID != nil {
		where = append(where, fmt.Sprintf("assignee_id = $%d", argIdx))
		args = append(args, *assigneeID)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", whereClause)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, status, priority, project_id, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, params.Limit, params.Offset())

	var tasks []model.Task
	if err := r.db.SelectContext(ctx, &tasks, query, args...); err != nil {
		return nil, 0, err
	}
	if tasks == nil {
		tasks = []model.Task{}
	}
	return tasks, total, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	query := `
		UPDATE tasks SET title = $1, description = $2, status = $3, priority = $4,
		assignee_id = $5, due_date = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at`
	return r.db.QueryRowContext(ctx, query,
		task.Title, task.Description, task.Status, task.Priority,
		task.AssigneeID, task.DueDate, task.ID,
	).Scan(&task.UpdatedAt)
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TaskRepository) GetStatsByProject(ctx context.Context, projectID uuid.UUID) (*model.ProjectStats, error) {
	// Count by status
	var statusCounts []model.StatusCount
	statusQuery := `SELECT status, COUNT(*) as count FROM tasks WHERE project_id = $1 GROUP BY status`
	if err := r.db.SelectContext(ctx, &statusCounts, statusQuery, projectID); err != nil {
		return nil, err
	}

	byStatus := map[model.TaskStatus]int{
		model.TaskStatusTodo:       0,
		model.TaskStatusInProgress: 0,
		model.TaskStatusDone:       0,
	}
	for _, sc := range statusCounts {
		byStatus[sc.Status] = sc.Count
	}

	// Count by assignee
	var assigneeCounts []model.AssigneeCount
	assigneeQuery := `
		SELECT t.assignee_id as user_id, COALESCE(u.name, 'Unassigned') as name, COUNT(*) as count
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		WHERE t.project_id = $1
		GROUP BY t.assignee_id, u.name
		ORDER BY count DESC`
	if err := r.db.SelectContext(ctx, &assigneeCounts, assigneeQuery, projectID); err != nil {
		return nil, err
	}
	if assigneeCounts == nil {
		assigneeCounts = []model.AssigneeCount{}
	}

	return &model.ProjectStats{
		ByStatus:   byStatus,
		ByAssignee: assigneeCounts,
	}, nil
}
