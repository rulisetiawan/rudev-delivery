package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleWahaWebhookLocationParsing(t *testing.T) {
	payload := WahaWebhookPayload{
		Event:   "message",
		Session: "default",
	}
	payload.Payload.From = "6281234567890@c.us"
	payload.Payload.Type = "location"
	payload.Payload.Data.Lat = -6.917464
	payload.Payload.Data.Lng = 107.619125

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/waha/webhook", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handleWahaWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	bodyStr := w.Body.String()
	if bodyStr != `{"status":"success"}` {
		t.Errorf("Expected success JSON response, got %s", bodyStr)
	}
}

func TestHandleWahaWebhookInvalidMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/waha/webhook", nil)
	w := httptest.NewRecorder()

	handleWahaWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 Method Not Allowed, got %d", resp.StatusCode)
	}
}
