package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

func TestDrawerObjectsAdvertiseImageOutputModalities(t *testing.T) {
	fresh(t)
	drawingRelay(t, "drawers", "image")

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set(provider.DrawersHeader, "1")
	res := httptest.NewRecorder()
	New().Handler().ServeHTTP(res, req)
	var list struct {
		Data []struct {
			ID               string   `json:"id"`
			OutputModalities []string `json:"output_modalities"`
		}
	}
	if err := json.Unmarshal(res.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	for _, model := range list.Data {
		if model.ID != "drawers/image" {
			continue
		}
		if !slices.Equal(model.OutputModalities, []string{"image"}) {
			t.Fatalf("output_modalities = %#v, want [image]", model.OutputModalities)
		}
		return
	}
	t.Fatal("drawer model was not listed")
}
