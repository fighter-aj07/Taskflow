package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/taskflow/backend/internal/config"
	"github.com/taskflow/backend/internal/handler"
	"github.com/taskflow/backend/internal/middleware"
	"github.com/taskflow/backend/internal/repository"
	"github.com/taskflow/backend/internal/service"
)

var (
	testDB     *sqlx.DB
	testRouter *chi.Mux
)

func TestMain(m *testing.M) {
	cfg := config.Config{
		DBHost:     getTestEnv("POSTGRES_HOST", "localhost"),
		DBPort:     getTestEnv("POSTGRES_PORT", "5432"),
		DBUser:     getTestEnv("POSTGRES_USER", "taskflow"),
		DBPassword: getTestEnv("POSTGRES_PASSWORD", "taskflow_secret"),
		DBName:     getTestEnv("POSTGRES_DB", "taskflow_test"),
		APIPort:    "8080",
		JWTSecret:  "test-secret-for-integration-tests",
		BcryptCost: 4, // Low cost for fast tests
	}

	var err error
	testDB, err = sqlx.Connect("postgres", cfg.DatabaseURL())
	if err != nil {
		fmt.Printf("Failed to connect to test DB: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	// Wire up
	userRepo := repository.NewUserRepository(testDB)
	projectRepo := repository.NewProjectRepository(testDB)
	taskRepo := repository.NewTaskRepository(testDB)

	authService := service.NewAuthService(userRepo, cfg)
	projectService := service.NewProjectService(projectRepo)
	taskService := service.NewTaskService(taskRepo, projectRepo)

	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService, taskService)
	taskHandler := handler.NewTaskHandler(taskService)
	authMW := middleware.NewAuthMiddleware(authService)

	testRouter = chi.NewRouter()
	testRouter.Use(chimiddleware.Recoverer)

	testRouter.Post("/auth/register", authHandler.Register)
	testRouter.Post("/auth/login", authHandler.Login)

	testRouter.Group(func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Get("/projects", projectHandler.List)
		r.Post("/projects", projectHandler.Create)
		r.Get("/projects/{id}", projectHandler.Get)
		r.Patch("/projects/{id}", projectHandler.Update)
		r.Delete("/projects/{id}", projectHandler.Delete)
		r.Get("/projects/{id}/stats", projectHandler.Stats)
		r.Get("/projects/{id}/tasks", taskHandler.List)
		r.Post("/projects/{id}/tasks", taskHandler.Create)
		r.Patch("/tasks/{id}", taskHandler.Update)
		r.Delete("/tasks/{id}", taskHandler.Delete)
	})

	os.Exit(m.Run())
}

func cleanDB(t *testing.T) {
	t.Helper()
	testDB.MustExec("TRUNCATE tasks, projects, users CASCADE")
}

func doRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)
	return rr
}

func registerUser(name, email, password string) (string, string) {
	body := map[string]string{"name": name, "email": email, "password": password}
	rr := doRequest("POST", "/auth/register", body, "")
	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	token := data["token"].(string)
	user := data["user"].(map[string]interface{})
	userID := user["id"].(string)
	return token, userID
}

func getTestEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
