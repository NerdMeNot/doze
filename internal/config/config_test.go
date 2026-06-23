package config

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/nerdmenot/doze/internal/engine"
)

func init() { engine.Register(fakeDriver{}) }

// fakeDriver is a minimal engine used to exercise the generic config parser
// without depending on a real engine package.
type fakeDriver struct{}

type fakeConfig struct {
	Color string
	Ref   string // populated from a cross-instance reference, e.g. fake.a.name
}

func (fakeDriver) Type() string { return "fake" }
func (fakeDriver) Resolve(context.Context, engine.VersionSpec, engine.Platform, engine.Locker, engine.Fetcher) (engine.Toolchain, error) {
	return engine.Toolchain{}, nil
}
func (fakeDriver) Provision(context.Context, engine.Instance, engine.Toolchain) error { return nil }
func (fakeDriver) Provisioned(string) bool                                            { return false }
func (fakeDriver) Spawn(context.Context, engine.Instance, engine.Toolchain) (engine.Process, error) {
	return nil, nil
}
func (fakeDriver) WaitReady(context.Context, engine.Instance, engine.Toolchain, engine.Process) error {
	return nil
}
func (fakeDriver) BackendSocket(string, int) string                             { return "" }
func (fakeDriver) ConnString(engine.Instance, engine.Endpoint) (string, string) { return "", "" }
func (fakeDriver) DecodeConfig(body hcl.Body, ctx *hcl.EvalContext, _ string) (engine.EngineConfig, error) {
	var s struct {
		Color string `hcl:"color,optional"`
		Ref   string `hcl:"ref,optional"`
	}
	if d := gohcl.DecodeBody(body, ctx, &s); d.HasErrors() {
		return nil, d
	}
	return &fakeConfig{Color: s.Color, Ref: s.Ref}, nil
}

func TestParseRootAndInstances(t *testing.T) {
	src := `
listen = "127.0.0.1:7000"
defaults { idle_timeout = "90s" }

fake "a" {
  version = 16
  color   = "red"
}
fake "b" {
  version = "17.2"
}
`
	cfg, err := Parse([]byte(src), "doze.hcl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Listen != "127.0.0.1:7000" {
		t.Errorf("listen = %q", cfg.Listen)
	}
	if cfg.Defaults.IdleTimeout != 90*time.Second {
		t.Errorf("idle_timeout = %s", cfg.Defaults.IdleTimeout)
	}
	if len(cfg.Instances) != 2 {
		t.Fatalf("got %d instances, want 2", len(cfg.Instances))
	}
	a := cfg.Lookup("a")
	if a == nil || a.Type != "fake" || a.Version != "16" {
		t.Fatalf("a = %+v", a)
	}
	fc, ok := a.Spec.(*fakeConfig)
	if !ok || fc.Color != "red" {
		t.Errorf("a.Spec = %+v", a.Spec)
	}
	if b := cfg.Lookup("b"); b == nil || b.Version != "17.2" {
		t.Errorf("b = %+v", b)
	}
}

func TestVersionRequired(t *testing.T) {
	_, err := Parse([]byte(`fake "x" {}`), "doze.hcl")
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Errorf("expected missing-version error, got %v", err)
	}
}

func TestDuplicateInstance(t *testing.T) {
	src := `
fake "dup" { version = 1 }
fake "dup" { version = 1 }
`
	_, err := Parse([]byte(src), "doze.hcl")
	if err == nil || !strings.Contains(err.Error(), "already declared") {
		t.Errorf("expected duplicate error, got %v", err)
	}
}

func TestUnknownEngine(t *testing.T) {
	_, err := Parse([]byte(`mysql "x" { version = 8 }`), "doze.hcl")
	if err == nil || !strings.Contains(err.Error(), "mysql") {
		t.Errorf("expected unknown-engine error mentioning mysql, got %v", err)
	}
}

func TestDefaultsApplied(t *testing.T) {
	cfg, err := Parse([]byte(`fake "x" { version = 1 }`), "doze.hcl")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Listen != DefaultListen {
		t.Errorf("default listen = %q", cfg.Listen)
	}
	if cfg.Defaults.IdleTimeout != DefaultIdleTimeout {
		t.Errorf("default idle_timeout = %s", cfg.Defaults.IdleTimeout)
	}
	if cfg.Home == "" || cfg.DataDir == "" {
		t.Error("home and data_dir should default")
	}
}

func TestBadDuration(t *testing.T) {
	_, err := Parse([]byte(`defaults { idle_timeout = "not-a-duration" }`), "doze.hcl")
	if err == nil || !strings.Contains(err.Error(), "idle_timeout") {
		t.Errorf("expected duration error, got %v", err)
	}
}

func TestUnknownKey(t *testing.T) {
	src := `
defaults {
  bogus_key = "oops"
}
`
	_, err := Parse([]byte(src), "doze.hcl")
	if err == nil || !strings.Contains(err.Error(), "bogus_key") {
		t.Errorf("expected unknown-key error, got %v", err)
	}
}

func TestReferenceResolvesAndBuildsEdge(t *testing.T) {
	src := `
fake "a" { version = 1 }
fake "b" {
  version = 1
  ref     = fake.a.name
}
`
	cfg, err := Parse([]byte(src), "doze.hcl")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	b := cfg.Lookup("b")
	if b == nil {
		t.Fatal("instance b missing")
	}
	// The reference resolves to a value...
	if got := b.Spec.(*fakeConfig).Ref; got != "a" {
		t.Errorf("ref = %q, want %q", got, "a")
	}
	// ...and the dependency edge is derived from it.
	if len(b.Deps) != 1 || b.Deps[0] != "a" {
		t.Errorf("b.Deps = %v, want [a]", b.Deps)
	}
	// The referenced instance has no deps.
	if a := cfg.Lookup("a"); len(a.Deps) != 0 {
		t.Errorf("a.Deps = %v, want none", a.Deps)
	}
}

func TestReferenceToUndeclaredInstance(t *testing.T) {
	src := `
fake "b" {
  version = 1
  ref     = fake.ghost.name
}
`
	_, err := Parse([]byte(src), "doze.hcl")
	if err == nil || !strings.Contains(err.Error(), "undeclared instance") || !strings.Contains(err.Error(), "ghost") {
		t.Errorf("expected undeclared-instance error, got %v", err)
	}
}

func TestReferenceCycleDetected(t *testing.T) {
	src := `
fake "x" {
  version = 1
  ref     = fake.y.name
}
fake "y" {
  version = 1
  ref     = fake.x.name
}
`
	_, err := Parse([]byte(src), "doze.hcl")
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected cycle error, got %v", err)
	}
}

func TestFunctionsAvailableInExpressions(t *testing.T) {
	src := `
fake "a" {
  version = 1
  color   = upper("red")
}
`
	cfg, err := Parse([]byte(src), "doze.hcl")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := cfg.Lookup("a").Spec.(*fakeConfig).Color; got != "RED" {
		t.Errorf("color = %q, want RED (upper applied)", got)
	}
}
