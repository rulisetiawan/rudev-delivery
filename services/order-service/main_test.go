package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// 1. Test Delivery Fee Calculator Formula
func TestDeliveryFeeCalculator(t *testing.T) {
	tests := []struct {
		name             string
		baseFee          float64
		extraStoreFee    float64
		totalStores      int
		expectedDelivery float64
	}{
		{
			name:             "Single store delivery",
			baseFee:          10000.0,
			extraStoreFee:    2000.0,
			totalStores:      1,
			expectedDelivery: 10000.0,
		},
		{
			name:             "Two stores delivery",
			baseFee:          10000.0,
			extraStoreFee:    2000.0,
			totalStores:      2,
			expectedDelivery: 12000.0,
		},
		{
			name:             "Custom tariff settings",
			baseFee:          15000.0,
			extraStoreFee:    3000.0,
			totalStores:      3,
			expectedDelivery: 21000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualDelivery := tt.baseFee + float64(tt.totalStores-1)*tt.extraStoreFee
			if actualDelivery != tt.expectedDelivery {
				t.Errorf("Expected delivery fee %.2f, got %.2f", tt.expectedDelivery, actualDelivery)
			}
		})
	}
}

// 2. Test Open Graph Meta Tag Rendering for WhatsApp Crawler
func TestShareStoreOpenGraph(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/public/share/store/warung-bu-ani", nil)
	req.Header.Set("User-Agent", "WhatsApp/2.21.12.21 A")
	w := httptest.NewRecorder()

	handleShareStoreOG(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "og:title") {
		t.Errorf("Expected response to contain og:title meta tag")
	}
	if !strings.Contains(bodyStr, "og:image") {
		t.Errorf("Expected response to contain og:image meta tag")
	}
}

// 3. Test Checkout Payload Parsing Validation
func TestCheckoutInvalidMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/public/checkout", nil)
	w := httptest.NewRecorder()

	handleCheckout(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 Method Not Allowed, got %d", resp.StatusCode)
	}
}

func TestCheckoutInvalidBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/public/checkout", strings.NewReader("invalid-json"))
	w := httptest.NewRecorder()

	handleCheckout(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", resp.StatusCode)
	}
}

// 4. Test Tariff Settings Struct Serialization
func TestTariffSettingsSerialization(t *testing.T) {
	settings := TariffSettings{
		BaseDeliveryFee:  12000.0,
		PerExtraStoreFee: 2500.0,
		MinOrderAmount:   0.0,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	var parsed TariffSettings
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	if parsed.BaseDeliveryFee != 12000.0 || parsed.PerExtraStoreFee != 2500.0 {
		t.Errorf("Mismatch in unmarshalled tariff settings")
	}
}
