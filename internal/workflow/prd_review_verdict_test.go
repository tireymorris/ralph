package workflow

import "testing"

func TestParsePRDReviewVerdict(t *testing.T) {
	t.Run("valid verdict", func(t *testing.T) {
		v := parsePRDReviewVerdict([]byte(`{"approved": true, "summary": "looks good"}`))
		if !v.Approved {
			t.Error("expected Approved to be true")
		}
		if v.Summary != "looks good" {
			t.Errorf("Summary = %q, want %q", v.Summary, "looks good")
		}
	})

	t.Run("malformed json returns zero-value verdict", func(t *testing.T) {
		v := parsePRDReviewVerdict([]byte(`not json`))
		if v != (PRDReviewVerdict{}) {
			t.Errorf("expected zero-value verdict, got %+v", v)
		}
	})

	t.Run("empty data returns zero-value verdict", func(t *testing.T) {
		v := parsePRDReviewVerdict(nil)
		if v != (PRDReviewVerdict{}) {
			t.Errorf("expected zero-value verdict, got %+v", v)
		}
	})
}
