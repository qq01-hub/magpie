package gateway

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

func TestGroupImageSizeRefusalFailsOverWithoutCoolingMember(t *testing.T) {
	for _, status := range []int{400, 413} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			fresh(t)
			const imagePrefix = "data:image/png;base64,"
			image := func(size int) string { return imagePrefix + strings.Repeat("A", size) }
			request := func(src string) string {
				return `{"model":"group/g","messages":[{"role":"user","content":[{"type":"text","text":"read"},{"type":"image_url","image_url":{"url":"` + src + `"}}]}]}`
			}
			a := &scripted{replies: []reply{
				{200, "", `{"choices":[{"message":{"role":"assistant","content":"from-a"},"finish_reason":"stop"}]}`},
				{status, "", `{"error":{"message":"model \"m\" rejects images larger than 1048576 bytes (got 1400000 bytes)"}}`},
				{200, "", `{"choices":[{"message":{"role":"assistant","content":"from-a"},"finish_reason":"stop"}]}`},
			}}
			b := &scripted{replies: []reply{{200, "", `{"choices":[{"message":{"role":"assistant","content":"from-b"},"finish_reason":"stop"}]}`}}}
			scriptedOn(t, "a", provider.Chat, a)
			scriptedOn(t, "b", provider.Chat, b)
			if err := provider.SaveGroup(provider.Group{Name: "G", Members: []string{"a/m", "b/m"}, Routing: provider.Ordered}); err != nil {
				t.Fatal(err)
			}
			s := New()
			if code, body := postAs(t, s, "", request(image(8))); code != 200 || !strings.Contains(body, "from-a") || a.n != 1 || b.n != 0 {
				t.Fatalf("small image: %d %s (a %d, b %d)", code, body, a.n, b.n)
			}
			if code, body := postAs(t, s, "", request(image(1<<20))); code != 200 || !strings.Contains(body, "from-b") || a.n != 2 || b.n != 1 {
				t.Fatalf("large image did not reach sibling: %d %s (a %d, b %d)", code, body, a.n, b.n)
			}
			tries := lastRoute(s).Tries
			if len(tries) != 2 || tries[0].Status != status || tries[0].Rest != nil && tries[0].Rest.By != "off" {
				t.Fatalf("image-size limitation did not fail over without a cooldown: %+v", tries)
			}
			if code, body := postAs(t, s, "", request(image(8))); code != 200 || !strings.Contains(body, "from-a") || a.n != 3 || b.n != 1 {
				t.Fatalf("member did not serve a later small image: %d %s (a %d, b %d)", code, body, a.n, b.n)
			}
		})
	}
}

func TestImageSizeRefusalRetryClassification(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   bool
		limit  bool
	}{
		{"image bytes over vendor limit", 400, `{"error":{"message":"model rejects images larger than 1048576 bytes"}}`, true, true},
		{"image payload limit", 413, `{"error":{"message":"Maximum image size is 1 MB"}}`, true, true},
		{"image size exceeds byte limit", 400, `{"error":{"message":"Image size exceeds 2 MiB"}}`, true, true},
		{"ordinary invalid request", 400, `{"error":{"message":"messages: field required"}}`, false, false},
		{"invalid image format", 400, `{"error":{"message":"unsupported image format"}}`, false, false},
		{"format and unrelated cap in separate sentences", 400, `{"error":{"message":"unsupported image format. Request exceeds the maximum size"}}`, false, false},
		{"image count limit", 400, `{"error":{"message":"Maximum images per request is 4"}}`, false, false},
		{"image maximum count wording", 400, `{"error":{"message":"images exceed maximum count of 4"}}`, false, false},
		{"image rate limit wording", 400, `{"error":{"message":"images exceeded request rate limit"}}`, true, false},
		{"image with context overflow", 413, `{"error":{"message":"Image size is 2 MiB. Request exceeds the model context length"}}`, false, false},
		{"generic upstream payload cap", 413, `{"error":{"type":"request_too_large","message":"Request exceeds the maximum size"}}`, false, false},
		{"context overflow", 413, `{"error":{"message":"request exceeds the model context length"}}`, false, false},
		{"image wording at unauthorized status", 401, `{"error":{"message":"images larger than 1048576 bytes"}}`, true, false},
		{"image wording at forbidden status", 403, `{"error":{"message":"images larger than 1048576 bytes"}}`, true, false},
		{"image wording at rate limit status", 429, `{"error":{"message":"images larger than 1048576 bytes"}}`, true, false},
		{"wrong status", 422, `{"error":{"message":"images larger than 1048576 bytes"}}`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := retryable(tc.status, []byte(tc.body)); got != tc.want {
				t.Fatalf("retryable(%d, %s) = %v, want %v", tc.status, tc.body, got, tc.want)
			}
			if got := imageSizeLimitRefused(tc.status, []byte(tc.body)); got != tc.limit {
				t.Fatalf("imageSizeLimitRefused(%d, %s) = %v, want %v", tc.status, tc.body, got, tc.limit)
			}
		})
	}
}
