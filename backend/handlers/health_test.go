package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	// ایجاد درخواست تست
	httptest.NewRequest(http.MethodGet, "/health", nil)

	// ایجاد پاسخ‌دهنده تست
	w := httptest.NewRecorder()

	// فرض می‌کنیم یک handler به نام HealthCheck داریم
	// HealthCheck(w, req)

	// بررسی وضعیت پاسخ
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}

	// بررسی محتوای پاسخ
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
}
