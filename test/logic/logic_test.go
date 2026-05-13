package logic_test

import (
	"testing"

	"github.com/franklee-labs/celero-go/condition"
	"github.com/franklee-labs/celero-go/core"
	"github.com/franklee-labs/celero-go/logic"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// makeEQ builds a compiled EqualCondition with a string value.
func makeEQ(id, field, value string) core.Condition {
	cfg := core.NewConditionConfig(id, id)
	c := condition.NewEqualCondition(field, value, core.ValueTypeString, cfg)
	if err := c.Compile(); err != nil {
		panic("makeEQ compile: " + err.Error())
	}
	return c
}

// makeGT builds a compiled GT compare condition.
func makeGT(id, field, value string) core.Condition {
	cfg := core.NewConditionConfig(id, id)
	c := condition.NewCompareCondition(field, value, "GT", core.ValueTypeNumber, cfg)
	if err := c.Compile(); err != nil {
		panic("makeGT compile: " + err.Error())
	}
	return c
}

// ─── AND ─────────────────────────────────────────────────────────────────────

func TestAND_ValidateNoChildren(t *testing.T) {
	a := logic.NewAND()
	v := a.Validate()
	if v.Valid {
		t.Fatal("AND with no children should be invalid")
	}
}

func TestAND_ValidateOneChild(t *testing.T) {
	a := logic.NewAND()
	a.SetChildren([]core.Node{makeEQ("c1", "name", "alice")})
	v := a.Validate()
	if !v.Valid {
		t.Fatalf("AND with one child should be valid: %s", v.Message)
	}
}

func TestAND_RelKind(t *testing.T) {
	a := logic.NewAND()
	if a.RelKind() != core.RelationAnd {
		t.Fatal("expected RelationAnd")
	}
}

// Resolve single condition → one path with one condition.
func TestAND_ResolveSingleCondition(t *testing.T) {
	c := makeEQ("c1", "name", "alice")
	a := logic.NewAND()
	a.SetChildren([]core.Node{c})
	rel, err := a.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	pg := rel.PathGroup()
	if pg.Size() != 1 {
		t.Fatalf("expected 1 path, got %d", pg.Size())
	}
	if pg.Get(0).Size() != 1 {
		t.Fatalf("expected 1 condition in path, got %d", pg.Get(0).Size())
	}
}

// AND(c1, c2) → one path with both conditions.
func TestAND_ResolveTwoConditions(t *testing.T) {
	c1 := makeEQ("c1", "name", "alice")
	c2 := makeGT("c2", "age", "18")
	a := logic.NewAND()
	a.SetChildren([]core.Node{c1, c2})
	rel, err := a.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	pg := rel.PathGroup()
	if pg.Size() != 1 {
		t.Fatalf("AND(c1,c2) should produce 1 path, got %d", pg.Size())
	}
	if pg.Get(0).Size() != 2 {
		t.Fatalf("path should have 2 conditions, got %d", pg.Get(0).Size())
	}
}

// AND(c1, OR(c2, c3)) → two paths: [c1,c2] and [c1,c3].
func TestAND_ResolveWithNestedOR(t *testing.T) {
	c1 := makeEQ("c1", "role", "admin")
	c2 := makeEQ("c2", "dept", "eng")
	c3 := makeEQ("c3", "dept", "ops")

	or := logic.NewOR()
	or.SetChildren([]core.Node{c2, c3})

	a := logic.NewAND()
	a.SetChildren([]core.Node{c1, or})

	rel, err := a.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	pg := rel.PathGroup()
	if pg.Size() != 2 {
		t.Fatalf("AND(c1,OR(c2,c3)) should produce 2 paths, got %d", pg.Size())
	}
	for i := 0; i < pg.Size(); i++ {
		if pg.Get(i).Size() != 2 {
			t.Fatalf("path %d should have 2 conditions (c1 + one of c2/c3), got %d",
				i, pg.Get(i).Size())
		}
	}
}

// AND.Negate() → OR of negated children.
func TestAND_Negate(t *testing.T) {
	c1 := makeEQ("c1", "name", "alice")
	c2 := makeEQ("c2", "city", "NY")
	a := logic.NewAND()
	a.SetChildren([]core.Node{c1, c2})
	neg, err := a.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	if neg.RelKind() != core.RelationOr {
		t.Fatalf("Negate(AND) should produce OR, got %v", neg.RelKind())
	}
}

// ─── OR ──────────────────────────────────────────────────────────────────────

func TestOR_ValidateLessThanTwoChildren(t *testing.T) {
	o := logic.NewOR()
	v := o.Validate()
	if v.Valid {
		t.Fatal("OR with 0 children should be invalid")
	}
	o.SetChildren([]core.Node{makeEQ("c1", "name", "alice")})
	v = o.Validate()
	if v.Valid {
		t.Fatal("OR with 1 child should be invalid")
	}
}

func TestOR_ValidateTwoChildren(t *testing.T) {
	o := logic.NewOR()
	o.SetChildren([]core.Node{makeEQ("c1", "name", "alice"), makeEQ("c2", "name", "bob")})
	v := o.Validate()
	if !v.Valid {
		t.Fatalf("OR with 2 children should be valid: %s", v.Message)
	}
}

func TestOR_RelKind(t *testing.T) {
	o := logic.NewOR()
	if o.RelKind() != core.RelationOr {
		t.Fatal("expected RelationOr")
	}
}

// OR(c1, c2) → two paths, each with one condition.
func TestOR_ResolveTwoConditions(t *testing.T) {
	c1 := makeEQ("c1", "name", "alice")
	c2 := makeEQ("c2", "name", "bob")
	o := logic.NewOR()
	o.SetChildren([]core.Node{c1, c2})
	rel, err := o.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	pg := rel.PathGroup()
	if pg.Size() != 2 {
		t.Fatalf("OR(c1,c2) should produce 2 paths, got %d", pg.Size())
	}
	for i := 0; i < pg.Size(); i++ {
		if pg.Get(i).Size() != 1 {
			t.Fatalf("each OR path should have 1 condition, got %d", pg.Get(i).Size())
		}
	}
}

// OR(AND(c1,c2), AND(c3,c4)) → two paths each with 2 conditions.
func TestOR_ResolveNestedANDs(t *testing.T) {
	c1, c2, c3, c4 := makeEQ("c1", "a", "1"), makeEQ("c2", "b", "2"),
		makeEQ("c3", "c", "3"), makeEQ("c4", "d", "4")

	and1 := logic.NewAND()
	and1.SetChildren([]core.Node{c1, c2})
	and2 := logic.NewAND()
	and2.SetChildren([]core.Node{c3, c4})

	o := logic.NewOR()
	o.SetChildren([]core.Node{and1, and2})
	rel, err := o.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	pg := rel.PathGroup()
	if pg.Size() != 2 {
		t.Fatalf("OR(AND,AND) should produce 2 paths, got %d", pg.Size())
	}
	for i := 0; i < pg.Size(); i++ {
		if pg.Get(i).Size() != 2 {
			t.Fatalf("path %d should have 2 conditions, got %d", i, pg.Get(i).Size())
		}
	}
}

// OR.Negate() → AND of negated children (De Morgan's).
func TestOR_Negate(t *testing.T) {
	c1 := makeEQ("c1", "name", "alice")
	c2 := makeEQ("c2", "name", "bob")
	o := logic.NewOR()
	o.SetChildren([]core.Node{c1, c2})
	neg, err := o.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	if neg.RelKind() != core.RelationAnd {
		t.Fatalf("Negate(OR) should produce AND, got %v", neg.RelKind())
	}
}

// ─── NOT ─────────────────────────────────────────────────────────────────────

func TestNOT_ValidateNoChildren(t *testing.T) {
	n := logic.NewNOT()
	v := n.Validate()
	if v.Valid {
		t.Fatal("NOT with 0 children should be invalid")
	}
}

func TestNOT_ValidateTwoChildren(t *testing.T) {
	n := logic.NewNOT()
	n.SetChildren([]core.Node{makeEQ("c1", "a", "1"), makeEQ("c2", "b", "2")})
	v := n.Validate()
	if v.Valid {
		t.Fatal("NOT with 2 children should be invalid")
	}
}

func TestNOT_ValidateOneChild(t *testing.T) {
	n := logic.NewNOT()
	n.SetChildren([]core.Node{makeEQ("c1", "name", "alice")})
	v := n.Validate()
	if !v.Valid {
		t.Fatalf("NOT with 1 child should be valid: %s", v.Message)
	}
}

func TestNOT_RelKind(t *testing.T) {
	n := logic.NewNOT()
	if n.RelKind() != core.RelationNot {
		t.Fatal("expected RelationNot")
	}
}

// NOT(EQ) resolves to a single path with the negated condition.
func TestNOT_ResolveCondition(t *testing.T) {
	c := makeEQ("c1", "name", "alice")
	n := logic.NewNOT()
	n.SetChildren([]core.Node{c})
	rel, err := n.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	pg := rel.PathGroup()
	if pg.Size() != 1 {
		t.Fatalf("NOT(cond) should produce 1 path, got %d", pg.Size())
	}
	if pg.Get(0).Size() != 1 {
		t.Fatalf("path should contain 1 (negated) condition, got %d", pg.Get(0).Size())
	}
}

// NOT(condition).Resolve() → 1 path with the negated condition.
// (NOT of a nested relation is handled by the engine via ConditionNode.Transform;
// testing NOT(AND/OR) directly at the logic layer is out of scope here.)
func TestNOT_ResolveConditionNegatesValue(t *testing.T) {
	c := makeEQ("c1", "role", "admin")
	n := logic.NewNOT()
	n.SetChildren([]core.Node{c})
	rel, err := n.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	pg := rel.PathGroup()
	if pg == nil || pg.Size() != 1 {
		t.Fatalf("NOT(cond) should produce 1 path, got %v", pg)
	}
	if pg.Get(0).Size() != 1 {
		t.Fatalf("path should have 1 (negated) condition, got %d", pg.Get(0).Size())
	}
}

// NOT.Negate() with a condition child returns AND wrapping the original condition.
func TestNOT_NegateConditionChild(t *testing.T) {
	c := makeEQ("c1", "name", "alice")
	n := logic.NewNOT()
	n.SetChildren([]core.Node{c})
	neg, err := n.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	if neg.RelKind() != core.RelationAnd {
		t.Fatalf("Negate(NOT(cond)) should return AND, got %v", neg.RelKind())
	}
}

// NOT.Negate() with a relation child returns that relation directly.
func TestNOT_NegateRelationChild(t *testing.T) {
	c1 := makeEQ("c1", "a", "1")
	c2 := makeEQ("c2", "b", "2")
	inner := logic.NewOR()
	inner.SetChildren([]core.Node{c1, c2})

	n := logic.NewNOT()
	n.SetChildren([]core.Node{inner})
	neg, err := n.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	if neg.RelKind() != core.RelationOr {
		t.Fatalf("Negate(NOT(OR)) should return OR, got %v", neg.RelKind())
	}
}

// ─── ValidateAll ─────────────────────────────────────────────────────────────

func TestValidateAll_ValidTree(t *testing.T) {
	c1 := makeEQ("c1", "name", "alice")
	c2 := makeEQ("c2", "city", "NY")
	a := logic.NewAND()
	a.SetChildren([]core.Node{c1, c2})
	v := core.ValidateAll(a)
	if !v.Valid {
		t.Fatalf("valid tree should pass ValidateAll: %s", v.Message)
	}
}

func TestValidateAll_EmptyAND(t *testing.T) {
	a := logic.NewAND()
	v := core.ValidateAll(a)
	if v.Valid {
		t.Fatal("AND with no children should fail ValidateAll")
	}
}

func TestValidateAll_InvalidConditionInTree(t *testing.T) {
	// CEL condition with empty expression fails validation
	badCond := condition.NewCELCondition("", core.NewConditionConfig("bad", "bad"))
	a := logic.NewAND()
	a.SetChildren([]core.Node{badCond})
	v := core.ValidateAll(a)
	if v.Valid {
		t.Fatal("tree with invalid condition should fail ValidateAll")
	}
}
