package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRegister_Success(t *testing.T) {
	cleanDB(t)
	body := map[string]string{
		"name": "Jane Doe", "email": "jane@example.com", "password": "secret123",
	}
	rr := doRequest("POST", "/auth/register", body, "")

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["token"] == nil || data["token"] == "" {
		t.Fatal("expected token in response")
	}
	user := data["user"].(map[string]interface{})
	if user["email"] != "jane@example.com" {
		t.Fatalf("expected email jane@example.com, got %s", user["email"])
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	cleanDB(t)
	body := map[string]string{
		"name": "Jane", "email": "dupe@example.com", "password": "secret123",
	}
	doRequest("POST", "/auth/register", body, "")
	rr := doRequest("POST", "/auth/register", body, "")

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestLogin_Success(t *testing.T) {
	cleanDB(t)
	regBody := map[string]string{
		"name": "Jane", "email": "jane@example.com", "password": "secret123",
	}
	doRequest("POST", "/auth/register", regBody, "")

	loginBody := map[string]string{
		"email": "jane@example.com", "password": "secret123",
	}
	rr := doRequest("POST", "/auth/login", loginBody, "")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if data["token"] == nil || data["token"] == "" {
		t.Fatal("expected token in login response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	cleanDB(t)
	regBody := map[string]string{
		"name": "Jane", "email": "jane@example.com", "password": "secret123",
	}
	doRequest("POST", "/auth/register", regBody, "")

	loginBody := map[string]string{
		"email": "jane@example.com", "password": "wrong",
	}
	rr := doRequest("POST", "/auth/login", loginBody, "")

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}
