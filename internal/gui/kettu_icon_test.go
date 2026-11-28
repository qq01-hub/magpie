package gui

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKettuClientIconServed(t *testing.T) {
	w := httptest.NewRecorder()
	http.FileServer(http.FS(staticFS())).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/icons/kettu.png", nil))
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("Kettu embedded icon response = %d %q", w.Code, w.Header().Get("Content-Type"))
	}
	config, err := png.DecodeConfig(bytes.NewReader(w.Body.Bytes()))
	if err != nil || config.Width != 64 || config.Height != 64 {
		t.Fatalf("Kettu icon dimensions = %dx%d, error = %v", config.Width, config.Height, err)
	}
	if icon, ok := pageIconOf("kettu"); !ok || icon.Mono {
		t.Fatalf("Kettu Usage icon missing or treated as a mask: %+v %v", icon, ok)
	}
}
