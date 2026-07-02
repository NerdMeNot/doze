package runtime

import (
	"testing"

	"github.com/doze-dev/doze-sdk/engine"
)

// rawSpec mimics plugin.RawSpec: an opaque-bytes spec exposing Raw().
type rawSpec struct{ b []byte }

func (r rawSpec) Raw() []byte { return r.b }

func TestSpecFingerprint(t *testing.T) {
	// No spec → empty fingerprint (falls back to marker-presence semantics).
	if fp := specFingerprint(engine.Instance{Spec: nil}); fp != "" {
		t.Errorf("nil spec: got %q, want empty", fp)
	}
	// Empty opaque bytes → empty fingerprint.
	if fp := specFingerprint(engine.Instance{Spec: rawSpec{}}); fp != "" {
		t.Errorf("empty raw: got %q, want empty", fp)
	}

	// Same opaque bytes → same fingerprint; different bytes → different.
	a := specFingerprint(engine.Instance{Spec: rawSpec{b: []byte("role app")}})
	a2 := specFingerprint(engine.Instance{Spec: rawSpec{b: []byte("role app")}})
	b := specFingerprint(engine.Instance{Spec: rawSpec{b: []byte("role app, role api")}})
	if a == "" || a != a2 {
		t.Errorf("stability: %q vs %q", a, a2)
	}
	if a == b {
		t.Errorf("drift not detected: both %q", a)
	}

	// In-tree (non-Raw) spec uses the gob fallback and is still stable + sensitive.
	type cfg struct{ Roles []string }
	c1 := specFingerprint(engine.Instance{Spec: &cfg{Roles: []string{"app"}}})
	c1b := specFingerprint(engine.Instance{Spec: &cfg{Roles: []string{"app"}}})
	c2 := specFingerprint(engine.Instance{Spec: &cfg{Roles: []string{"app", "api"}}})
	if c1 == "" || c1 != c1b {
		t.Errorf("gob fallback stability: %q vs %q", c1, c1b)
	}
	if c1 == c2 {
		t.Errorf("gob fallback drift not detected: both %q", c1)
	}
}
