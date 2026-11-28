package gateway

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInternLMSafetyRefusal(t *testing.T) {
	// Standard InternLM error body
	raw := `{"message":"The request contains sensitive content. Please modify your input and try again.","type":"content_check_error","code":"content_check_rejected"}`

	// 1. Test policyRefusal with raw error
	if note, ok := policyRefusal([]byte(raw)); !ok {
		t.Errorf("expected raw InternLM error to be recognized as policyRefusal, got ok=%v, note=%s", ok, note)
	} else if !strings.Contains(note, "sensitive") && !strings.Contains(note, "content_filter") {
		t.Errorf("expected note to contain sensitive/content_filter, got %s", note)
	}

	// 2. Test policyRefusal inside standard error envelope
	enveloped, _ := json.Marshal(map[string]any{
		"error": map[string]any{
			"message": "The request contains sensitive content. Please modify your input and try again.",
			"type":    "content_check_error",
			"code":    "content_check_rejected",
		},
	})
	if note, ok := policyRefusal(enveloped); !ok {
		t.Errorf("expected enveloped InternLM error to be recognized as policyRefusal, got ok=%v, note=%s", ok, note)
	}

	// 3. Test stream event parsing
	event := `data: {"error":{"message":"The request contains sensitive content. Please modify your input and try again.","type":"content_check_error","code":"content_check_rejected"}}`
	kind, status, msg := streamEvent([]byte(event))
	if kind != eventRefusalError {
		t.Errorf("expected kind == eventRefusalError (%d), got %d", eventRefusalError, kind)
	}
	if status != refusedStatus {
		t.Errorf("expected status == refusedStatus (%d), got %d", refusedStatus, status)
	}
	if !strings.Contains(msg, "sensitive") && !strings.Contains(msg, "content_filter") {
		t.Errorf("expected msg to contain refusal note, got %s", msg)
	}
}
