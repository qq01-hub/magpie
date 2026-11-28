package gateway

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSocketPostBodyOwnershipAtRetentionLimit(t *testing.T) {
	rs := responsesSocket{o: &socketOrigin{req: httptest.NewRequest(http.MethodGet, "/v1/responses", nil)}}
	for _, size := range []int{64, promptBodyMax, promptBodyMax + 1} {
		body := bytes.Repeat([]byte(" "), size)
		copy(body, `{"input":[]}`)
		req := rs.request(context.Background(), body)
		got, status, err := readBoundedRequestBody(httptest.NewRecorder(), req, requestLimits{body: int64(size)})
		if err != nil || status != 0 || !bytes.Equal(got, body) {
			t.Fatalf("%d-byte turn changed: status %d, %v", size, status, err)
		}
		if shared := &got[0] == &body[0]; shared != (size > promptBodyMax) {
			t.Errorf("%d-byte turn shares its socket backing storage: %v", size, shared)
		}
	}
}

func TestSocketPostBodyLimitDoesNotTrustDeclaredLength(t *testing.T) {
	body := bytes.Repeat([]byte(" "), promptBodyMax+1)
	copy(body, `{"input":[]}`)
	rs := responsesSocket{o: &socketOrigin{req: httptest.NewRequest(http.MethodGet, "/v1/responses", nil)}}
	req := rs.request(context.Background(), body)
	req.ContentLength = 1
	if _, status, err := readBoundedRequestBody(httptest.NewRecorder(), req, requestLimits{body: 1024}); status != http.StatusRequestEntityTooLarge || err == nil {
		t.Fatalf("large socket body with a short declared length: status %d, %v", status, err)
	}
}
