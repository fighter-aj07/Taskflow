package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Enums

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone:
		return true
	}
	return false
}

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

func (p TaskPriority) IsValid() bool {
	switch p {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh:
		return true
	}
	return false
}

// Entities

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Project struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Task struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Title       string         `json:"title" db:"title"`
	Description *string        `json:"description" db:"description"`
	Status      TaskStatus     `json:"status" db:"status"`
	Priority    TaskPriority   `json:"priority" db:"priority"`
	ProjectID   uuid.UUID      `json:"project_id" db:"project_id"`
	AssigneeID  *uuid.UUID     `json:"assignee_id" db:"assignee_id"`
	CreatedBy   uuid.UUID      `json:"created_by" db:"created_by"`
	DueDate     sql.NullTime   `json:"due_date" db:"due_date"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// Pagination

type PaginationParams struct {
	Page  int
	Limit int
}

func NewPaginationParams(page, limit int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return PaginationParams{Page: page, Limit: limit}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

type PaginatedResult[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPagination(params PaginationParams, total int) Pagination {
	totalPages := total / params.Limit
	if total%params.Limit > 0 {
		totalPages++
	}
	return Pagination{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// Stats (bonus)

type StatusCount struct {
	Status TaskStatus `json:"status" db:"status"`
	Count  int        `json:"count" db:"count"`
}

type AssigneeCount struct {
	UserID *uuid.UUID `json:"user_id" db:"user_id"`
	Name   string     `json:"name" db:"name"`
	Count  int        `json:"count" db:"count"`
}

type ProjectStats struct {
	ByStatus   map[TaskStatus]int `json:"by_status"`
	ByAssignee []AssigneeCount    `json:"by_assignee"`
}

// Project detail (includes tasks)

type ProjectDetail struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	Tasks       []Task    `json:"tasks"`
}

func NewProjectDetail(p *Project, tasks []Task) ProjectDetail {
	if tasks == nil {
		tasks = []Task{}
	}
	return ProjectDetail{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		OwnerID:     p.OwnerID,
		CreatedAt:   p.CreatedAt,
		Tasks:       tasks,
	}
}

// Request/Response DTOs

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateTaskRequest struct {
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	Priority    TaskPriority `json:"priority"`
	AssigneeID  *uuid.UUID   `json:"assignee_id"`
	DueDate     *string      `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	Status      *TaskStatus   `json:"status"`
	Priority    *TaskPriority `json:"priority"`
	AssigneeID  *uuid.UUID    `json:"assignee_id"`
	DueDate     *string       `json:"due_date"`
}
