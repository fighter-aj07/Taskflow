package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateTask_AndFilterByStatus(t *testing.T) {
	cleanDB(t)
	token, _ := registerUser("Jane", "jane@example.com", "secret123")

	// Create project
	projBody := map[string]string{"name": "Test Project"}
	rr := doRequest("POST", "/projects", projBody, token)
	var projResp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&projResp)
	projectID := projResp["data"].(map[string]interface{})["id"].(string)

	// Create tasks with different statuses
	task1 := map[string]interface{}{"title": "Task 1", "priority": "high"}
	doRequest("POST", fmt.Sprintf("/projects/%s/tasks", projectID), task1, token)

	task2 := map[string]interface{}{"title": "Task 2", "priority": "low"}
	rr = doRequest("POST", fmt.Sprintf("/projects/%s/tasks", projectID), task2, token)
	var task2Resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&task2Resp)
	task2ID := task2Resp["data"].(map[string]interface{})["id"].(string)

	// Update task2 to done
	updateBody := map[string]string{"status": "done"}
	doRequest("PATCH", fmt.Sprintf("/tasks/%s", task2ID), updateBody, token)

	// Filter by status=todo
	rr = doRequest("GET", fmt.Sprintf("/projects/%s/tasks?status=todo", projectID), nil, token)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var listResp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&listResp)
	tasks := listResp["data"].([]interface{})
	if len(tasks) != 1 {
		t.Fatalf("expected 1 todo task, got %d", len(tasks))
	}
}

func TestDeleteTask_Unauthorized(t *testing.T) {
	cleanDB(t)
	ownerToken, _ := registerUser("Owner", "owner@example.com", "secret123")
	otherToken, _ := registerUser("Other", "other@example.com", "secret123")

	// Owner creates project and task
	projBody := map[string]string{"name": "Owner Project"}
	rr := doRequest("POST", "/projects", projBody, ownerToken)
	var projResp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&projResp)
	projectID := projResp["data"].(map[string]interface{})["id"].(string)

	taskBody := map[string]string{"title": "Owner Task"}
	rr = doRequest("POST", fmt.Sprintf("/projects/%s/tasks", projectID), taskBody, ownerToken)
	var taskResp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&taskResp)
	taskID := taskResp["data"].(map[string]interface{})["id"].(string)

	// Other tries to delete (not owner, not creator)
	rr = doRequest("DELETE", fmt.Sprintf("/tasks/%s", taskID), nil, otherToken)

	// Other has no access to the project at all, so this should be 403 or 404
	if rr.Code != http.StatusForbidden && rr.Code != http.StatusNotFound {
		t.Fatalf("expected 403 or 404, got %d: %s", rr.Code, rr.Body.String())
	}
}
