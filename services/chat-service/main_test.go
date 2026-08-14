package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatMessageSerialization(t *testing.T) {
	msg := ChatMessage{
		OrderID:     8821,
		SenderID:    1,
		SenderRole:  "customer",
		MessageText: "Halo kurir, posisi di mana?",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal ChatMessage: %v", err)
	}

	var parsed ChatMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal ChatMessage: %v", err)
	}

	if parsed.MessageText != msg.MessageText {
		t.Errorf("Expected message text %s, got %s", msg.MessageText, parsed.MessageText)
	}
}

func TestHandleChatHistory(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/chat/history?order_id=8821", nil)
	w := httptest.NewRecorder()

	handleChatHistory(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
}
