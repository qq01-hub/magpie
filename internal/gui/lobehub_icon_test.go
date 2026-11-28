package gui

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLobeHubClientIconServed(t *testing.T) {
	w := httptest.NewRecorder()
	http.FileServer(http.FS(staticFS())).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/icons/lobehub-color.svg", nil))
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("LobeHub embedded asset response = %d %q", w.Code, w.Header().Get("Content-Type"))
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("<svg")) || !bytes.Contains(w.Body.Bytes(), []byte("<title>LobeHub</title>")) {
		t.Fatal("LobeHub response is not the logo SVG")
	}
	if icon, ok := pageIconOf("lobehub-color"); !ok || icon.Mono {
		t.Fatalf("LobeHub usage-page icon missing or treated as a mask: %+v %v", icon, ok)
	}
}
