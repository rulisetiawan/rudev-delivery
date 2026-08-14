package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleChargeQRISPayload(t *testing.T) {
	reqBody := ChargeReq{
		OrderID:     8821,
		GrossAmount: 25000,
		PaymentType: "QRIS",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/payments/charge", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handleChargeQRIS(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d", resp.StatusCode)
	}

	var res map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&res)

	if res["payment_type"] != "QRIS" {
		t.Errorf("Expected payment_type QRIS, got %v", res["payment_type"])
	}
}

func TestHandleMidtransWebhookInvalidMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/payments/webhook", nil)
	w := httptest.NewRecorder()

	handleMidtransWebhook(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 Method Not Allowed, got %d", resp.StatusCode)
	}
}
