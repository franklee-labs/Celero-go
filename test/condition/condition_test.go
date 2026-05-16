package condition_test

import (
	"strings"
	"testing"

	"github.com/franklee-labs/celero-go/condition"
	"github.com/franklee-labs/celero-go/core"
	celtypes "github.com/google/cel-go/common/types"
)

// ─── mock EvalContext ────────────────────────────────────────────────────────

type mockCtx struct {
	params     map[string]interface{}
	evalParams map[string]interface{}
	cache      map[string]core.EvalResult
	absEnabled bool
}

func newCtx(params map[string]interface{}) *mockCtx {
	return &mockCtx{
		params:     params,
		evalParams: params,
		cache:      make(map[string]core.EvalResult),
		absEnabled: true,
	}
}

func (m *mockCtx) Params() map[string]interface{}      { return m.params }
func (m *mockCtx) IsConditionResultCacheEnabled() bool { return false }
func (m *mockCtx) IsAbsenceEnabled() bool              { return m.absEnabled }
func (m *mockCtx) IsReportEnabled() bool               { return false }
func (m *mockCtx) BuildEvalParams(p map[string]interface{}) error {
	merged := make(map[string]interface{}, len(m.params)+len(p))
	for k, v := range m.params {
		merged[k] = v
	}
	for k, v := range p {
		if !strings.HasPrefix(k, "_") {
			return nil // silently skip; real engine validates this
		}
		merged[k] = v
	}
	m.evalParams = merged
	return nil
}
func (m *mockCtx) EvalParams() map[string]interface{} { return m.evalParams }
func (m *mockCtx) RuleContext() core.RuleContext      { return nil }
func (m *mockCtx) SetConditionEvalResult(id string, r core.EvalResult) {
	m.cache[id] = r
}
func (m *mockCtx) GetConditionEvalResult(id string) (core.EvalResult, bool) {
	r, ok := m.cache[id]
	return r, ok
}

// helper — creates a ConditionConfig with given id/name
func cfg(id, name string) *core.ConditionConfig {
	return core.NewConditionConfig(id, name)
}

// helper — compiles and asserts no error
func mustCompile(t *testing.T, cond interface{ Compile() error }) {
	t.Helper()
	if err := cond.Compile(); err != nil {
		t.Fatalf("Compile() error: %v", err)
	}
}

// ─── EqualCondition ──────────────────────────────────────────────────────────

func TestEqualString_Match(t *testing.T) {
	c := condition.NewEqualCondition("name", "alice", core.ValueTypeString, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"name": "alice"})
	c.BeforeEvaluate(ctx)
	ok, absent, err := c.Evaluate(ctx)
	if err != nil || absent || !ok {
		t.Fatalf("expected true, got ok=%v absent=%v err=%v", ok, absent, err)
	}
}

func TestEqualString_NoMatch(t *testing.T) {
	c := condition.NewEqualCondition("name", "alice", core.ValueTypeString, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"name": "bob"})
	c.BeforeEvaluate(ctx)
	ok, absent, err := c.Evaluate(ctx)
	if err != nil || absent || ok {
		t.Fatalf("expected false, got ok=%v absent=%v err=%v", ok, absent, err)
	}
}

func TestEqualString_MissingField(t *testing.T) {
	c := condition.NewEqualCondition("name", "alice", core.ValueTypeString, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{})
	c.BeforeEvaluate(ctx)
	_, absent, _ := c.Evaluate(ctx)
	if !absent {
		t.Fatal("expected absent=true when field is missing")
	}
}

func TestEqualNumber_IntegerMatch(t *testing.T) {
	c := condition.NewEqualCondition("age", "18", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(18)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestEqualNumber_FloatMatch(t *testing.T) {
	c := condition.NewEqualCondition("score", "9.5", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"score": 9.5})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestEqualNumber_InvalidValidate(t *testing.T) {
	c := condition.NewEqualCondition("age", "not-a-number", core.ValueTypeNumber, cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("expected invalid when value is not a number")
	}
}

func TestEqualBool_TrueMatch(t *testing.T) {
	c := condition.NewEqualCondition("active", "true", core.ValueTypeBool, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"active": true})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestEqualBool_FalseMatch(t *testing.T) {
	c := condition.NewEqualCondition("active", "false", core.ValueTypeBool, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"active": false})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestEqualExpression_Match(t *testing.T) {
	// "a == b" compares two runtime variables
	c := condition.NewEqualCondition("x", "y", core.ValueTypeExpression, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"x": int64(5), "y": int64(5)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestEqualNegate_ProducesNotEqual(t *testing.T) {
	c := condition.NewEqualCondition("name", "alice", core.ValueTypeString, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	// negated condition must evaluate to false when field equals value
	ctx := newCtx(map[string]interface{}{"name": "alice"})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("negated equal should be false when name==alice, got ok=%v err=%v", ok, err)
	}
}

// ─── NotEqualCondition ───────────────────────────────────────────────────────

func TestNotEqualString_Match(t *testing.T) {
	c := condition.NewNotEqualCondition("name", "alice", core.ValueTypeString, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"name": "bob"})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestNotEqualString_NoMatch(t *testing.T) {
	c := condition.NewNotEqualCondition("name", "alice", core.ValueTypeString, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"name": "alice"})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("expected false, got ok=%v err=%v", ok, err)
	}
}

func TestNotEqualNumber(t *testing.T) {
	c := condition.NewNotEqualCondition("age", "18", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(20)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("20 != 18 should be true, got ok=%v err=%v", ok, err)
	}
}

func TestNotEqualBool(t *testing.T) {
	c := condition.NewNotEqualCondition("active", "true", core.ValueTypeBool, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"active": false})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("false != true should be true, got ok=%v err=%v", ok, err)
	}
}

func TestNotEqualNegate_ProducesEqual(t *testing.T) {
	c := condition.NewNotEqualCondition("name", "alice", core.ValueTypeString, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	// negated not-equal should be true when name==alice
	ctx := newCtx(map[string]interface{}{"name": "alice"})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("negated NEQ should be true when name==alice, got ok=%v err=%v", ok, err)
	}
}

// ─── CompareCondition ────────────────────────────────────────────────────────

func TestCompareGT_True(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "GT", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(20)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("20 > 18 should be true, got ok=%v err=%v", ok, err)
	}
}

func TestCompareGT_False(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "GT", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(10)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("10 > 18 should be false")
	}
}

func TestCompareGTE_Equal(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "GTE", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(18)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("18 >= 18 should be true")
	}
}

func TestCompareGTE_False(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "GTE", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(10)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("10 >= 18 should be false")
	}
}

func TestCompareLT_True(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "LT", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(10)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("10 < 18 should be true")
	}
}

func TestCompareLT_False(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "LT", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(20)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("20 < 18 should be false")
	}
}

func TestCompareLTE_Equal(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "LTE", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(18)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("18 <= 18 should be true")
	}
}

func TestCompareLTE_False(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "LTE", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(20)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("20 <= 18 should be false")
	}
}

func TestCompareFloat(t *testing.T) {
	c := condition.NewCompareCondition("score", "9.5", "GT", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"score": 9.9})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("9.9 > 9.5 should be true, got ok=%v err=%v", ok, err)
	}
}

func TestCompareExpression(t *testing.T) {
	c := condition.NewCompareCondition("x", "y", "GT", core.ValueTypeExpression, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"x": int64(10), "y": int64(5)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("x > y should be true, got ok=%v err=%v", ok, err)
	}
}

func TestCompareInvalidSign(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "INVALID", core.ValueTypeNumber, cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("expected invalid for unsupported sign")
	}
}

func TestCompareInvalidValueType(t *testing.T) {
	c := condition.NewCompareCondition("age", "alice", "GT", core.ValueTypeString, cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("expected invalid for String value type in Compare")
	}
}

func TestCompareNegateGT_ProducesLTE(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "GT", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"age": int64(18)})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	// NOT(age > 18) at age=18 → 18 <= 18 → true
	if err != nil || !ok {
		t.Fatalf("NOT(age>18) at 18 should be true (LTE), got ok=%v err=%v", ok, err)
	}
}

// ─── InCondition ─────────────────────────────────────────────────────────────

func TestIn_ListHit(t *testing.T) {
	c := condition.NewInCondition("status", `["active","pending"]`, core.ValueTypeList, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"status": "active"})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("'active' in list should be true, got ok=%v err=%v", ok, err)
	}
}

func TestIn_ListMiss(t *testing.T) {
	c := condition.NewInCondition("status", `["active","pending"]`, core.ValueTypeList, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"status": "closed"})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("'closed' in list should be false")
	}
}

func TestIn_EmptyList_InvalidValidation(t *testing.T) {
	c := condition.NewInCondition("status", `[]`, core.ValueTypeList, cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("empty list should be invalid")
	}
}

func TestIn_InvalidJSON_InvalidValidation(t *testing.T) {
	c := condition.NewInCondition("status", `not-json`, core.ValueTypeList, cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("invalid JSON should be invalid")
	}
}

func TestIn_Expression(t *testing.T) {
	// "status in allowedList" - both runtime variables
	c := condition.NewInCondition("status", "allowedList", core.ValueTypeExpression, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"status": "active", "allowedList": []interface{}{"active", "pending"}})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("expression in should be true, got ok=%v err=%v", ok, err)
	}
}

func TestIn_NegateProducesNotIn(t *testing.T) {
	c := condition.NewInCondition("status", `["active"]`, core.ValueTypeList, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"status": "active"})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	// NOT IN: 'active' IS in list → negated → false
	if err != nil || ok {
		t.Fatalf("NOT IN should be false when value is in list, got ok=%v err=%v", ok, err)
	}
}

// ─── NotInCondition ──────────────────────────────────────────────────────────

func TestNotIn_ListHit(t *testing.T) {
	c := condition.NewNotInCondition("status", `["active","pending"]`, core.ValueTypeList, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"status": "closed"})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("'closed' NOT IN list should be true, got ok=%v err=%v", ok, err)
	}
}

func TestNotIn_ListMiss(t *testing.T) {
	c := condition.NewNotInCondition("status", `["active"]`, core.ValueTypeList, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"status": "active"})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("'active' NOT IN list should be false")
	}
}

func TestNotIn_NegateProducesIn(t *testing.T) {
	c := condition.NewNotInCondition("status", `["active"]`, core.ValueTypeList, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"status": "active"})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("negated NOT-IN should be true when value is in list, got ok=%v err=%v", ok, err)
	}
}

// ─── ExistsCondition ─────────────────────────────────────────────────────────

func TestExists_FieldPresent(t *testing.T) {
	c := condition.NewExistsCondition("params.name", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"params": map[string]interface{}{"name": "alice"}})
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("field present → exists should be true, got ok=%v err=%v", ok, err)
	}
}

func TestExists_FieldAbsent(t *testing.T) {
	c := condition.NewExistsCondition("params.age", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"params": map[string]interface{}{"name": "alice"}})
	ok, absent, err := c.Evaluate(ctx)
	// has() on missing key returns false; ExistsCondition suppresses absent
	_ = err
	if ok {
		t.Fatalf("missing field → exists should be false, got ok=%v absent=%v", ok, absent)
	}
}

func TestExists_EmptyField_InvalidValidation(t *testing.T) {
	c := condition.NewExistsCondition("", cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("empty field should be invalid")
	}
}

func TestExists_NegateProducesAbsent(t *testing.T) {
	c := condition.NewExistsCondition("params.age", cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	// AbsentCondition: missing key → true
	ctx := newCtx(map[string]interface{}{"params": map[string]interface{}{"name": "alice"}})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("absent should be true when key missing, got ok=%v err=%v", ok, err)
	}
}

// ─── AbsentCondition ─────────────────────────────────────────────────────────

func TestAbsent_FieldMissing_True(t *testing.T) {
	c := condition.NewAbsentCondition("params.age", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"params": map[string]interface{}{"name": "alice"}})
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("absent field should return true, got ok=%v err=%v", ok, err)
	}
}

func TestAbsent_FieldPresent_False(t *testing.T) {
	c := condition.NewAbsentCondition("params.name", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"params": map[string]interface{}{"name": "alice"}})
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("present field should return false for absent condition")
	}
}

func TestAbsent_EmptyField_InvalidValidation(t *testing.T) {
	c := condition.NewAbsentCondition("", cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("empty field should be invalid")
	}
}

func TestAbsent_NegateProducesExists(t *testing.T) {
	c := condition.NewAbsentCondition("params.name", cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"params": map[string]interface{}{"name": "alice"}})
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("negated absent (exists) should be true when field present, got ok=%v err=%v", ok, err)
	}
}

// ─── RegexpCondition ─────────────────────────────────────────────────────────

func TestRegexp_Match(t *testing.T) {
	c := condition.NewRegexpCondition("email", `^[a-z]+@example\.com$`, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"email": "alice@example.com"})
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("expected match, got ok=%v err=%v", ok, err)
	}
}

func TestRegexp_NoMatch(t *testing.T) {
	c := condition.NewRegexpCondition("email", `^[a-z]+@example\.com$`, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"email": "alice@other.com"})
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("expected no match")
	}
}

func TestRegexp_EmptyPattern_InvalidValidation(t *testing.T) {
	c := condition.NewRegexpCondition("email", "", cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("empty regexp should be invalid")
	}
}

func TestRegexp_InvalidPattern_InvalidValidation(t *testing.T) {
	c := condition.NewRegexpCondition("email", `[invalid(`, cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("invalid regexp should be invalid")
	}
}

func TestRegexp_NegateProducesNonMatch(t *testing.T) {
	c := condition.NewRegexpCondition("email", `^alice`, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"email": "alice@test.com"})
	ok, _, err := neg.Evaluate(ctx)
	// matches alice → negated → false
	if err != nil || ok {
		t.Fatalf("negated regexp should be false when pattern matches, got ok=%v err=%v", ok, err)
	}
}

func TestNegateRegexp_Negate_ReturnsOriginal(t *testing.T) {
	orig := condition.NewRegexpCondition("email", `^alice`, cfg("c1", ""))
	mustCompile(t, orig)
	neg, _ := orig.Negate()
	// calling Negate on NegateRegexp should return orig
	restored, err := neg.Negate()
	if err != nil {
		t.Fatalf("double Negate() error: %v", err)
	}
	if restored.Expression() != orig.Expression() {
		t.Fatalf("double negate should restore original expression: got %q want %q",
			restored.Expression(), orig.Expression())
	}
}

// ─── CELCondition ────────────────────────────────────────────────────────────

func TestCEL_TrueExpression(t *testing.T) {
	c := condition.NewCELCondition("age >= 18", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(20)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("age>=18 at 20 should be true, got ok=%v err=%v", ok, err)
	}
}

func TestCEL_FalseExpression(t *testing.T) {
	c := condition.NewCELCondition("age >= 18", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"age": int64(10)})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("age>=18 at 10 should be false")
	}
}

func TestCEL_EmptyExpression_InvalidValidation(t *testing.T) {
	c := condition.NewCELCondition("", cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("empty CEL expression should be invalid")
	}
}

func TestCEL_ComplexExpression(t *testing.T) {
	c := condition.NewCELCondition(`items.exists(x, x == "vip")`, cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"items": []interface{}{"regular", "vip"}})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("exists expression should be true, got ok=%v err=%v", ok, err)
	}
}

func TestCEL_NegateProducesInverse(t *testing.T) {
	c := condition.NewCELCondition("age >= 18", cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"age": int64(20)})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	// NOT(age>=18) at 20 → false
	if err != nil || ok {
		t.Fatalf("negated CEL should be false at age=20, got ok=%v err=%v", ok, err)
	}
}

func TestNegateCEL_Negate_ReturnsOriginal(t *testing.T) {
	orig := condition.NewCELCondition("age >= 18", cfg("c1", ""))
	mustCompile(t, orig)
	neg, _ := orig.Negate()
	restored, err := neg.Negate()
	if err != nil {
		t.Fatalf("double Negate() error: %v", err)
	}
	if restored.Expression() != orig.Expression() {
		t.Fatalf("double negate should restore original: got %q want %q",
			restored.Expression(), orig.Expression())
	}
}

// ─── IntersectCondition ──────────────────────────────────────────────────────

func TestIntersect_BothExpression_HasIntersection(t *testing.T) {
	c := condition.NewIntersectCondition("tags", "Expression", "allowed", "Expression", cfg("c1", ""))
	if v := c.Validate(); !v.Valid {
		t.Fatalf("Validate() should be valid: %s", v.Message)
	}
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{
		"tags":    []interface{}{"vip", "gold"},
		"allowed": []interface{}{"gold", "silver"},
	})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("intersect should be true (gold in both), got ok=%v err=%v", ok, err)
	}
}

func TestIntersect_BothExpression_NoIntersection(t *testing.T) {
	c := condition.NewIntersectCondition("tags", "Expression", "allowed", "Expression", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{
		"tags":    []interface{}{"vip"},
		"allowed": []interface{}{"gold", "silver"},
	})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("intersect should be false (no common elements)")
	}
}

func TestIntersect_ListType_HasIntersection(t *testing.T) {
	c := condition.NewIntersectCondition(`["a","b"]`, "List", "tags", "Expression", cfg("c1", ""))
	if v := c.Validate(); !v.Valid {
		t.Fatalf("Validate() should be valid: %s", v.Message)
	}
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"tags": []interface{}{"b", "c"}})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("intersect List×Expr should be true, got ok=%v err=%v", ok, err)
	}
}

func TestIntersect_InvalidValueType(t *testing.T) {
	c := condition.NewIntersectCondition("tags", "String", "allowed", "Expression", cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("invalid value type should fail validation")
	}
}

func TestIntersect_NegateProducesDisjoint(t *testing.T) {
	c := condition.NewIntersectCondition("tags", "Expression", "allowed", "Expression", cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	// disjoint: no intersection → true
	ctx := newCtx(map[string]interface{}{
		"tags":    []interface{}{"vip"},
		"allowed": []interface{}{"gold"},
	})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("negated intersect (disjoint) should be true, got ok=%v err=%v", ok, err)
	}
}

// ─── DisjointCondition ───────────────────────────────────────────────────────

func TestDisjoint_BothExpression_NoIntersection(t *testing.T) {
	c := condition.NewDisjointCondition("tags", "Expression", "blocked", "Expression", cfg("c1", ""))
	if v := c.Validate(); !v.Valid {
		t.Fatalf("Validate() should be valid: %s", v.Message)
	}
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{
		"tags":    []interface{}{"vip"},
		"blocked": []interface{}{"spam"},
	})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("disjoint should be true (no common), got ok=%v err=%v", ok, err)
	}
}

func TestDisjoint_BothExpression_HasIntersection(t *testing.T) {
	c := condition.NewDisjointCondition("tags", "Expression", "blocked", "Expression", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{
		"tags":    []interface{}{"spam", "vip"},
		"blocked": []interface{}{"spam"},
	})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("disjoint should be false (spam common)")
	}
}

func TestDisjoint_InvalidValueType(t *testing.T) {
	c := condition.NewDisjointCondition("tags", "Number", "blocked", "Expression", cfg("c1", ""))
	v := c.Validate()
	if v.Valid {
		t.Fatal("invalid value type should fail validation")
	}
}

func TestDisjoint_NegateProducesIntersect(t *testing.T) {
	c := condition.NewDisjointCondition("tags", "Expression", "blocked", "Expression", cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{
		"tags":    []interface{}{"spam"},
		"blocked": []interface{}{"spam"},
	})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("negated disjoint (intersect) should be true, got ok=%v err=%v", ok, err)
	}
}

// ─── CompareCondition: remaining negateSignStr branches ───────────────────────

// GTE → negate → LT: NOT(age >= 18) at age=10 → age < 18 → true
func TestCompareNegateGTE_ProducesLT(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "GTE", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"age": int64(10)})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("NOT(age>=18) at 10 should be true (LT), got ok=%v err=%v", ok, err)
	}
}

// LT → negate → GTE: NOT(age < 18) at age=18 → age >= 18 → true
func TestCompareNegateLT_ProducesGTE(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "LT", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"age": int64(18)})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("NOT(age<18) at 18 should be true (GTE), got ok=%v err=%v", ok, err)
	}
}

// LTE → negate → GT: NOT(age <= 18) at age=20 → age > 18 → true
func TestCompareNegateLTE_ProducesGT(t *testing.T) {
	c := condition.NewCompareCondition("age", "18", "LTE", core.ValueTypeNumber, cfg("c1", ""))
	mustCompile(t, c)
	neg, err := c.Negate()
	if err != nil {
		t.Fatalf("Negate() error: %v", err)
	}
	ctx := newCtx(map[string]interface{}{"age": int64(20)})
	neg.BeforeEvaluate(ctx)
	ok, _, err := neg.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("NOT(age<=18) at 20 should be true (GT), got ok=%v err=%v", ok, err)
	}
}

// ─── IsKeyAbsent: types.IsError(out) branch ───────────────────────────────────

// When CEL returns an error *value* (err==nil) containing "no such key",
// IsKeyAbsent must still detect absence via the types.IsError path.
func TestIsKeyAbsent_ErrorVal_NoSuchKey(t *testing.T) {
	errVal := celtypes.NewErr("no such key: missing")
	if !condition.IsKeyAbsent(errVal, nil) {
		t.Fatal("error val with 'no such key' should return absent=true")
	}
}

func TestIsKeyAbsent_ErrorVal_NoSuchAttribute(t *testing.T) {
	errVal := celtypes.NewErr("no such attribute: foo")
	if !condition.IsKeyAbsent(errVal, nil) {
		t.Fatal("error val with 'no such attribute' should return absent=true")
	}
}

func TestIsKeyAbsent_ErrorVal_OtherError(t *testing.T) {
	errVal := celtypes.NewErr("division by zero")
	if condition.IsKeyAbsent(errVal, nil) {
		t.Fatal("unrelated error val should return absent=false")
	}
}

// ─── DisjointCondition: List-type Compile branches ────────────────────────────

// Both fields are literal JSON lists → both builtin param branches in Compile are hit.
func TestDisjoint_BothList_Disjoint(t *testing.T) {
	c := condition.NewDisjointCondition(`["a","b"]`, "List", `["c","d"]`, "List", cfg("c1", ""))
	if v := c.Validate(); !v.Valid {
		t.Fatalf("Validate() should be valid: %s", v.Message)
	}
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("disjoint literal lists should be true, got ok=%v err=%v", ok, err)
	}
}

func TestDisjoint_BothList_HasIntersection(t *testing.T) {
	c := condition.NewDisjointCondition(`["a","b"]`, "List", `["b","c"]`, "List", cfg("c1", ""))
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || ok {
		t.Fatalf("intersecting literal lists should not be disjoint")
	}
}

// field1 is a literal list, field2 is a runtime expression variable.
func TestDisjoint_ListAndExpression_Disjoint(t *testing.T) {
	c := condition.NewDisjointCondition(`["x","y"]`, "List", "tags", "Expression", cfg("c1", ""))
	if v := c.Validate(); !v.Valid {
		t.Fatalf("Validate() should be valid: %s", v.Message)
	}
	mustCompile(t, c)
	ctx := newCtx(map[string]interface{}{"tags": []interface{}{"a", "b"}})
	c.BeforeEvaluate(ctx)
	ok, _, err := c.Evaluate(ctx)
	if err != nil || !ok {
		t.Fatalf("List×Expr disjoint should be true, got ok=%v err=%v", ok, err)
	}
}
