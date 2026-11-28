package catalog

import (
	"net/url"
	"slices"
	"strings"
)

func syntheticAPI(base string) bool {
	u, err := url.Parse(strings.TrimSpace(base))
	return err == nil && strings.EqualFold(u.Hostname(), "api.synthetic.new")
}

// syntheticFacts keeps the capabilities Synthetic lists, including for an
// Infinite alias whose underlying Hugging Face id isn't its request id
// (zent00, #7430). Its models docs describe Infinite as a separate service
// mode, so keep both ids and name that mode in its otherwise identical label.
func syntheticFacts(m *Model, r liveModel) {
	if p, ok := r.ReasoningParameters.(map[string]any); ok {
		if levels := levelsOf(p["efforts"]); len(levels) > 0 {
			m.Efforts = levels
		}
	}
	if slices.Contains(strs(r.SupportedFeatures), "reasoning") {
		m.Reasoning = true
	}
	if n, ok := r.OutputLength.(float64); ok && n > 0 {
		m.Output = int(n)
	}
	if strings.HasPrefix(m.ID, "infinite:") && !strings.Contains(strings.ToLower(m.Name), "infinite") {
		m.Name += " (Infinite)"
	}
}
