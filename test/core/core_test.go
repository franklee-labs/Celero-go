package core_test

import (
	"testing"

	"github.com/franklee-labs/celero-go/core"
)

// ─── EvalResult ──────────────────────────────────────────────────────────────

func TestEvalResultTrue(t *testing.T) {
	r := core.EvalResultTrue
	if !r.IsTrue() {
		t.Fatal("expected IsTrue")
	}
	if r.IsFalse() || r.IsIndeterminate() || r.IsUnknown() {
		t.Fatal("only IsTrue should be set")
	}
}

func TestEvalResultFalse(t *testing.T) {
	r := core.EvalResultFalse
	if !r.IsFalse() {
		t.Fatal("expected IsFalse")
	}
	if r.IsTrue() || r.IsIndeterminate() || r.IsUnknown() {
		t.Fatal("only IsFalse should be set")
	}
}

func TestEvalResultIndeterminate(t *testing.T) {
	r := core.EvalResultIndeterminate
	if !r.IsIndeterminate() {
		t.Fatal("expected IsIndeterminate")
	}
	if r.IsTrue() || r.IsFalse() || r.IsUnknown() {
		t.Fatal("only IsIndeterminate should be set")
	}
}

func TestEvalResultUnknown(t *testing.T) {
	r := core.EvalResultUnknown
	if !r.IsUnknown() {
		t.Fatal("expected IsUnknown")
	}
	if r.IsTrue() || r.IsFalse() || r.IsIndeterminate() {
		t.Fatal("only IsUnknown should be set")
	}
}

// ─── Validation ──────────────────────────────────────────────────────────────

func TestValidationOK(t *testing.T) {
	v := core.ValidationOK
	if !v.Valid {
		t.Fatal("expected Valid=true")
	}
	if v.Message != "" {
		t.Fatalf("expected empty message, got %q", v.Message)
	}
}

func TestInvalidReturnsMessage(t *testing.T) {
	v := core.Invalid("something wrong")
	if v.Valid {
		t.Fatal("expected Valid=false")
	}
	if v.Message != "something wrong" {
		t.Fatalf("unexpected message: %q", v.Message)
	}
}

// ─── ValueTypeFromString ─────────────────────────────────────────────────────

func TestValueTypeFromString(t *testing.T) {
	cases := []struct {
		input string
		want  core.ValueType
	}{
		{"String", core.ValueTypeString},
		{"Number", core.ValueTypeNumber},
		{"Boolean", core.ValueTypeBool},
		{"List", core.ValueTypeList},
		{"Expression", core.ValueTypeExpression},
		{"", core.ValueTypeInvalid},
		{"unknown", core.ValueTypeInvalid},
		{"string", core.ValueTypeInvalid}, // case-sensitive
	}
	for _, tc := range cases {
		got := core.ValueTypeFromString(tc.input)
		if got != tc.want {
			t.Errorf("ValueTypeFromString(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ─── ConditionConfig ─────────────────────────────────────────────────────────

func TestNewConditionConfigDefaults(t *testing.T) {
	cfg := core.NewConditionConfig("c1", "my condition")
	if cfg.ID() != "c1" {
		t.Fatalf("ID mismatch")
	}
	if cfg.Name() != "my condition" {
		t.Fatalf("Name mismatch")
	}
	if cfg.Priority() != core.PriorityDefault {
		t.Fatalf("default priority should be %d, got %d", core.PriorityDefault, cfg.Priority())
	}
	if cfg.Cacheable() {
		t.Fatal("default cacheable should be false")
	}
	if cfg.IgnoreAbsence() {
		t.Fatal("default ignoreAbsence should be false")
	}
}

func TestConditionConfigWithCacheable(t *testing.T) {
	cfg := core.NewConditionConfig("c1", "n").WithCacheable(true)
	if !cfg.Cacheable() {
		t.Fatal("expected Cacheable=true")
	}
}

func TestConditionConfigWithIgnoreAbsence(t *testing.T) {
	cfg := core.NewConditionConfig("c1", "n").WithIgnoreAbsence(true)
	if !cfg.IgnoreAbsence() {
		t.Fatal("expected IgnoreAbsence=true")
	}
}

func TestConditionConfigPriorityNormal(t *testing.T) {
	cfg := core.NewConditionConfig("c1", "n").WithPriority(10)
	if cfg.Priority() != 10 {
		t.Fatalf("expected priority 10, got %d", cfg.Priority())
	}
}

func TestConditionConfigPriorityClampedToHighest(t *testing.T) {
	// Values below PriorityHighest should be clamped up to PriorityHighest.
	cfg := core.NewConditionConfig("c1", "n").WithPriority(core.PriorityHighest - 1)
	if cfg.Priority() != core.PriorityHighest {
		t.Fatalf("expected priority clamped to %d, got %d", core.PriorityHighest, cfg.Priority())
	}
}

func TestConditionConfigPriorityClampedToLowest(t *testing.T) {
	cfg := core.NewConditionConfig("c1", "n").WithPriority(core.PriorityLowest + 1)
	if cfg.Priority() != core.PriorityLowest {
		t.Fatalf("expected priority clamped to %d, got %d", core.PriorityLowest, cfg.Priority())
	}
}

// ─── PathGroup ───────────────────────────────────────────────────────────────

func TestNewPathGroupEmpty(t *testing.T) {
	pg := core.NewPathGroup()
	if pg.Size() != 0 {
		t.Fatalf("expected empty PathGroup, got size %d", pg.Size())
	}
	if pg.Get(0) != nil {
		t.Fatal("Get on empty PathGroup should return nil")
	}
}

func TestPathGroupAppendAndSize(t *testing.T) {
	pg := core.NewPathGroup()
	p := core.NewPath(nil)
	pg.AppendPath(p)
	if pg.Size() != 1 {
		t.Fatalf("expected size 1, got %d", pg.Size())
	}
}

func TestPathGroupGetOutOfBounds(t *testing.T) {
	pg := core.NewPathGroup()
	if pg.Get(-1) != nil {
		t.Fatal("negative index should return nil")
	}
	if pg.Get(10) != nil {
		t.Fatal("out-of-range index should return nil")
	}
}

func TestPathGroupConstructorWithPaths(t *testing.T) {
	p1 := core.NewPath(nil)
	p2 := core.NewPath(nil)
	pg := core.NewPathGroup(p1, p2)
	if pg.Size() != 2 {
		t.Fatalf("expected size 2, got %d", pg.Size())
	}
	if pg.Get(0) != p1 || pg.Get(1) != p2 {
		t.Fatal("paths not in expected order")
	}
}

// ─── Path ────────────────────────────────────────────────────────────────────

func TestNewPathEmpty(t *testing.T) {
	p := core.NewPath(nil)
	if p.Size() != 0 {
		t.Fatalf("expected empty path, got %d", p.Size())
	}
	if p.Get(0) != nil {
		t.Fatal("Get on empty path should return nil")
	}
}

func TestPathGetOutOfBounds(t *testing.T) {
	p := core.NewPath(nil)
	if p.Get(-1) != nil || p.Get(99) != nil {
		t.Fatal("out-of-bounds Get should return nil")
	}
}
