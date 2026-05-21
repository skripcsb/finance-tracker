package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]int{"count": 3})

	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}

	var result map[string]int
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["count"] != 3 {
		t.Errorf("got count=%d, want 3", result["count"])
	}
}
