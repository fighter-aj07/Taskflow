package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateProject_Authenticated(t *testing.T) {
	cleanDB(t)
	token, _ := registerUser("Jane", "jane@example.com", "secret123")

	body := map[string]string{"name": "My Project", "description": "Test project"}
	rr := doRequest("POST", "/projects", body, token)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "My Project" {
		t.Fatalf("expected project name 'My Project', got %s", data["name"])
	}
}

func TestDeleteProject_NotOwner(t *testing.T) {
	cleanDB(t)
	ownerToken, _ := registerUser("Owner", "owner@example.com", "secret123")
	otherToken, _ := registerUser("Other", "other@example.com", "secret123")

	// Owner creates project
	body := map[string]string{"name": "Owner Project"}
	rr := doRequest("POST", "/projects", body, ownerToken)
	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	projectID := data["id"].(string)

	// Other tries to delete
	rr = doRequest("DELETE", fmt.Sprintf("/projects/%s", projectID), nil, otherToken)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
}
