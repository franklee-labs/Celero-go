package engine_test

import (
	"testing"

	"github.com/franklee-labs/celero-go/engine"
)

// ─── JSON helpers ─────────────────────────────────────────────────────────────

// condJSON produces the JSON for a single EQ condition node.
func condJSON(id, name, field, value, valueType string) string {
	return `{
		"id": "` + id + `",
		"name": "` + name + `",
		"type": "condition",
		"sign": "EQ",
		"cacheable": false,
		"ignoreAbsence": false,
		"properties": {
			"field": "` + field + `",
			"value": "` + value + `",
			"valueType": "` + valueType + `",
			"priority": 0
		}
	}`
}

// buildRule builds a CeleroRule from raw JSON, failing the test on error.
func buildRule(t *testing.T, id, json string) *engine.CeleroRule {
	t.Helper()
	b, err := engine.FromJSON(id, json)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	rule, err := b.Name(id).Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	return rule
}

// ─── FromJSON / Build errors ──────────────────────────────────────────────────

func TestFromJSON_InvalidJSON(t *testing.T) {
	_, err := engine.FromJSON("r1", `{not valid json}`)
	if err == nil {
		t.Fatal("invalid JSON should return error")
	}
}

func TestFromJSON_MissingType(t *testing.T) {
	_, err := engine.FromJSON("r1", `{"sign": "EQ"}`)
	if err == nil {
		t.Fatal("missing 'type' field should return error")
	}
}

func TestFromJSON_InvalidType(t *testing.T) {
	_, err := engine.FromJSON("r1", `{"type": "unknown"}`)
	if err == nil {
		t.Fatal("unknown type should return error")
	}
}

func TestBuild_MissingID(t *testing.T) {
	b, err := engine.FromJSON("", condJSON("c1", "n", "age", "18", "Number"))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	_, err = b.Build()
	if err == nil {
		t.Fatal("empty rule ID should return error")
	}
}

// ─── DefaultEngine — single condition ────────────────────────────────────────

func TestDefaultEngine_EqualString_True(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestDefaultEngine_EqualString_False(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "bob"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || ok {
		t.Fatalf("expected false, got ok=%v err=%v", ok, err)
	}
}

func TestDefaultEngine_EqualNumber_True(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "age check", "age", "18", "Number"))
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"age": int64(18)})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("expected true, got ok=%v err=%v", ok, err)
	}
}

func TestDefaultEngine_MissingField_AbsentTreatedAsFalse(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "age check", "age", "18", "Number"))
	eng := engine.NewDefaultEngine()
	// DefaultEngine does not enable absence detection → absent = false
	ctx := engine.NewRuleContext(map[string]interface{}{})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || ok {
		t.Fatalf("absent field in DefaultEngine should result in false, got ok=%v err=%v", ok, err)
	}
}

// ─── DefaultEngine — AND rule ─────────────────────────────────────────────────

func TestDefaultEngine_ANDRule_BothMatch(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "AND",
		"children": [
			{
				"id":"c1","name":"n","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"alice","valueType":"String","priority":0}
			},
			{
				"id":"c2","name":"n2","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"city","value":"NY","valueType":"String","priority":0}
			}
		]
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice", "city": "NY"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("AND both match should be true, got ok=%v err=%v", ok, err)
	}
}

func TestDefaultEngine_ANDRule_OneDoesNotMatch(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "AND",
		"children": [
			{
				"id":"c1","name":"n","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"alice","valueType":"String","priority":0}
			},
			{
				"id":"c2","name":"n2","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"city","value":"NY","valueType":"String","priority":0}
			}
		]
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice", "city": "LA"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || ok {
		t.Fatalf("AND one miss should be false, got ok=%v err=%v", ok, err)
	}
}

// ─── DefaultEngine — OR rule ──────────────────────────────────────────────────

func TestDefaultEngine_ORRule_FirstMatches(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "OR",
		"children": [
			{
				"id":"c1","name":"n","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"alice","valueType":"String","priority":0}
			},
			{
				"id":"c2","name":"n2","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"bob","valueType":"String","priority":0}
			}
		]
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("OR first match should be true, got ok=%v err=%v", ok, err)
	}
}

func TestDefaultEngine_ORRule_SecondMatches(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "OR",
		"children": [
			{
				"id":"c1","name":"n","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"alice","valueType":"String","priority":0}
			},
			{
				"id":"c2","name":"n2","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"bob","valueType":"String","priority":0}
			}
		]
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "bob"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("OR second match should be true, got ok=%v err=%v", ok, err)
	}
}

func TestDefaultEngine_ORRule_NoneMatch(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "OR",
		"children": [
			{
				"id":"c1","name":"n","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"alice","valueType":"String","priority":0}
			},
			{
				"id":"c2","name":"n2","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"bob","valueType":"String","priority":0}
			}
		]
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "charlie"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || ok {
		t.Fatalf("OR none match should be false")
	}
}

// ─── DefaultEngine — condition listeners ─────────────────────────────────────

type testConditionListener struct {
	events []*engine.ConditionEvent
	order  int
}

func (l *testConditionListener) OnResult(e *engine.ConditionEvent) {
	l.events = append(l.events, e)
}
func (l *testConditionListener) Order() int { return l.order }

func TestDefaultEngine_ConditionListener_Fires(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	eng := engine.NewDefaultEngine()
	listener := &testConditionListener{}
	eng.AddConditionListener(listener)

	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	eng.Evaluate(rule, ctx)

	if len(listener.events) == 0 {
		t.Fatal("condition listener should have received at least one event")
	}
	if listener.events[0].ConditionID() != "c1" {
		t.Fatalf("expected conditionID=c1, got %q", listener.events[0].ConditionID())
	}
	if !listener.events[0].Matched() {
		t.Fatal("condition should have matched")
	}
}

func TestDefaultEngine_ConditionListener_OrderSorted(t *testing.T) {
	eng := engine.NewDefaultEngine()
	l1 := &testConditionListener{order: 10}
	l2 := &testConditionListener{order: 1}
	l3 := &testConditionListener{order: 5}
	eng.AddConditionListener(l1)
	eng.AddConditionListener(l2)
	eng.AddConditionListener(l3)

	rule := buildRule(t, "r1", condJSON("c1", "n", "name", "alice", "String"))
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	eng.Evaluate(rule, ctx)

	// All listeners should fire; just verify they were added without panic
	for _, l := range []*testConditionListener{l1, l2, l3} {
		if len(l.events) == 0 {
			t.Fatal("all listeners should have fired")
		}
	}
}

// ─── DefaultEngine — rule listeners ──────────────────────────────────────────

type testRuleListener struct {
	events []*engine.RuleEvent
}

func (l *testRuleListener) OnResult(e *engine.RuleEvent) { l.events = append(l.events, e) }
func (l *testRuleListener) Order() int                   { return 0 }

func TestDefaultEngine_Evalutes_MultipleRules(t *testing.T) {
	rule1 := buildRule(t, "r1", condJSON("c1", "n", "name", "alice", "String"))
	rule2 := buildRule(t, "r2", condJSON("c2", "n", "name", "bob", "String"))

	eng := engine.NewDefaultEngine()
	listener := &testRuleListener{}
	eng.AddRuleListener(listener)

	// Evalutes fires rule listeners only on error; here no error so listeners
	// are not triggered — just verify no panic
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	eng.Evalutes([]*engine.CeleroRule{rule1, rule2}, ctx)
}

// ─── AdvancedEngine ───────────────────────────────────────────────────────────

func TestAdvancedEngine_EqualString_True(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	eng := engine.NewAdvancedEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	result, err := eng.Evaluate(rule, ctx)
	if err != nil || !result.IsTrue() {
		t.Fatalf("expected True, got result=%v err=%v", result, err)
	}
}

func TestAdvancedEngine_EqualString_False(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	eng := engine.NewAdvancedEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "bob"})
	result, err := eng.Evaluate(rule, ctx)
	if err != nil || !result.IsFalse() {
		t.Fatalf("expected False, got result=%v err=%v", result, err)
	}
}

func TestAdvancedEngine_MissingField_Indeterminate(t *testing.T) {
	raw := `{
		"id":"c1","name":"age check",
		"type": "condition",
		"sign": "EQ",
		"cacheable": false,
		"ignoreAbsence": false,
		"properties": {
			"field":"age","value":"18","valueType":"Number","priority":0
		}
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewAdvancedEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{})
	result, err := eng.Evaluate(rule, ctx)
	// AdvancedEngine with absence enabled → indeterminate when field absent
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsIndeterminate() {
		t.Fatalf("absent field should be Indeterminate, got %v", result)
	}
}

func TestAdvancedEngine_IgnoreAbsence_ReturnsFalse(t *testing.T) {
	raw := `{
		"id":"c1","name":"age check",
		"type": "condition",
		"sign": "EQ",
		"cacheable": true,
		"ignoreAbsence": true,
		"properties": {
			"field":"age","value":"18","valueType":"Number","priority":0
		}
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewAdvancedEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{})
	result, err := eng.Evaluate(rule, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ignoreAbsence=true → absence treated as false, not indeterminate
	if result.IsIndeterminate() {
		t.Fatal("ignoreAbsence=true should not produce Indeterminate")
	}
	if !result.IsFalse() {
		t.Fatalf("ignoreAbsence=true with absent field should be False, got %v", result)
	}
}

// OR where first path is absent but second path matches → True (not Indeterminate).
func TestAdvancedEngine_ORRule_OneAbsentOneMatch_ReturnsTrue(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "OR",
		"children": [
			{
				"id":"c1","name":"n","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"missing_field","value":"x","valueType":"String","priority":0}
			},
			{
				"id":"c2","name":"n2","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"alice","valueType":"String","priority":0}
			}
		]
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewAdvancedEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	result, err := eng.Evaluate(rule, ctx)
	if err != nil || !result.IsTrue() {
		t.Fatalf("one path absent, other matches → should be True, got %v err=%v", result, err)
	}
}

// ─── AdvancedEngine — listeners ───────────────────────────────────────────────

type testAdvCondListener struct {
	events []*engine.AdvancedConditionEvent
}

func (l *testAdvCondListener) OnResult(e *engine.AdvancedConditionEvent) {
	l.events = append(l.events, e)
}
func (l *testAdvCondListener) Order() int { return 0 }

func TestAdvancedEngine_ConditionListener_Fires(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	eng := engine.NewAdvancedEngine()
	listener := &testAdvCondListener{}
	eng.AddConditionListener(listener)

	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	eng.Evaluate(rule, ctx)

	if len(listener.events) == 0 {
		t.Fatal("advanced condition listener should have fired")
	}
	if !listener.events[0].Result().IsTrue() {
		t.Fatal("result should be True")
	}
}

// ─── RuleContext ──────────────────────────────────────────────────────────────

func TestRuleContext_GetParams(t *testing.T) {
	params := map[string]interface{}{"age": int64(18)}
	ctx := engine.NewRuleContext(params)
	got := ctx.GetParams()
	if got["age"] != int64(18) {
		t.Fatalf("expected age=18")
	}
}

func TestRuleContext_SetAndGetAttribute(t *testing.T) {
	ctx := engine.NewRuleContext(nil)
	ctx.SetAttribute("key", "value")
	v, ok := ctx.GetAttribute("key")
	if !ok || v != "value" {
		t.Fatalf("expected attribute key=value, got ok=%v v=%v", ok, v)
	}
}

func TestRuleContext_GetAttribute_Missing(t *testing.T) {
	ctx := engine.NewRuleContext(nil)
	_, ok := ctx.GetAttribute("missing")
	if ok {
		t.Fatal("missing attribute should return ok=false")
	}
}

func TestRuleContext_EnableReports(t *testing.T) {
	ctx := engine.NewRuleContext(nil)
	if ctx.IsReportEnabled() {
		t.Fatal("reports should be disabled by default")
	}
	ctx.EnableReports()
	if !ctx.IsReportEnabled() {
		t.Fatal("reports should be enabled after EnableReports()")
	}
}

// ─── Report / Routes ─────────────────────────────────────────────────────────

func TestDefaultEngine_Report_RecordsRoutes(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	ctx.EnableReports()

	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("evaluate error: ok=%v err=%v", ok, err)
	}

	reports := ctx.Reports()
	if len(reports) == 0 {
		t.Fatal("expected at least one report entry")
	}
	for _, report := range reports {
		if len(report.Routes()) == 0 {
			t.Fatal("report should have at least one route")
		}
	}
}

func TestDefaultEngine_Report_UnmatchedRecorded(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "bob"})
	ctx.EnableReports()

	eng.Evaluate(rule, ctx)

	for _, report := range ctx.Reports() {
		for _, route := range report.Routes() {
			if len(route.Unmatched()) > 0 {
				return // found unmatched condition in report
			}
		}
	}
	t.Fatal("unmatched conditions should appear in report when rule fails")
}

// ─── Solutions ────────────────────────────────────────────────────────────────

func TestCeleroRule_Solutions_SingleCondition(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "name check", "name", "alice", "String"))
	solutions := rule.Solutions()
	if solutions.Count() != 1 {
		t.Fatalf("single-condition rule should have 1 solution, got %d", solutions.Count())
	}
	sol, ok := solutions.SolutionAt(0)
	if !ok {
		t.Fatal("SolutionAt(0) returned ok=false")
	}
	if sol.Count() != 1 {
		t.Fatalf("solution should have 1 condition, got %d", sol.Count())
	}
	cond, ok := sol.ConditionAt(0)
	if !ok {
		t.Fatal("ConditionAt(0) returned ok=false")
	}
	if cond.ID() != "c1" {
		t.Fatalf("condition ID mismatch: got %q want c1", cond.ID())
	}
}

func TestCeleroRule_Solutions_ORRule_TwoSolutions(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "OR",
		"children": [
			{
				"id":"c1","name":"n1","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"alice","valueType":"String","priority":0}
			},
			{
				"id":"c2","name":"n2","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"name","value":"bob","valueType":"String","priority":0}
			}
		]
	}`
	rule := buildRule(t, "r1", raw)
	if rule.Solutions().Count() != 2 {
		t.Fatalf("OR rule should have 2 solutions, got %d", rule.Solutions().Count())
	}
}

func TestSolutions_SolutionAt_OutOfBounds(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "n", "name", "alice", "String"))
	_, ok := rule.Solutions().SolutionAt(-1)
	if ok {
		t.Fatal("SolutionAt(-1) should return ok=false")
	}
	_, ok = rule.Solutions().SolutionAt(99)
	if ok {
		t.Fatal("SolutionAt(99) should return ok=false")
	}
}

func TestSolution_ConditionAt_OutOfBounds(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "n", "name", "alice", "String"))
	sol, _ := rule.Solutions().SolutionAt(0)
	_, ok := sol.ConditionAt(-1)
	if ok {
		t.Fatal("ConditionAt(-1) should return ok=false")
	}
	_, ok = sol.ConditionAt(99)
	if ok {
		t.Fatal("ConditionAt(99) should return ok=false")
	}
}

// ─── Cacheable ────────────────────────────────────────────────────────────────

func TestCeleroRule_SetCacheable(t *testing.T) {
	rule := buildRule(t, "r1", condJSON("c1", "n", "name", "alice", "String"))
	rule.SetCacheable(true)
	if !rule.IsCacheable() {
		t.Fatal("SetCacheable(true) should make IsCacheable() return true")
	}
	rule.SetCacheable(false)
	if rule.IsCacheable() {
		t.Fatal("SetCacheable(false) should make IsCacheable() return false")
	}
}

// ─── DefaultEngine with cacheable conditions ──────────────────────────────────

func TestDefaultEngine_CacheableRule_EvaluatesCorrectly(t *testing.T) {
	raw := `{
		"id":"c1","name":"name check",
		"type": "condition",
		"sign": "EQ",
		"cacheable": true,
		"ignoreAbsence": false,
		"properties": {
			"field":"name","value":"alice","valueType":"String","priority":0
		}
	}`
	b, err := engine.FromJSON("r1", raw)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	rule, err := b.Name("r1").Cacheable(true).Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"name": "alice"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("cacheable rule should evaluate correctly, got ok=%v err=%v", ok, err)
	}
}

// ─── CompareCondition via FromJSON ────────────────────────────────────────────

func TestDefaultEngine_GTCondition_True(t *testing.T) {
	raw := `{
		"id":"c1","name":"age gt",
		"type": "condition",
		"sign": "GT",
		"cacheable": false,
		"ignoreAbsence": false,
		"properties": {
			"field":"age","value":"18","valueType":"Number","priority":0
		}
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"age": int64(20)})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("age 20 > 18 should be true, got ok=%v err=%v", ok, err)
	}
}

// ─── InCondition via FromJSON ─────────────────────────────────────────────────

func TestDefaultEngine_INCondition_Hit(t *testing.T) {
	raw := `{
		"id":"c1","name":"status in",
		"type": "condition",
		"sign": "IN",
		"cacheable": false,
		"ignoreAbsence": false,
		"properties": {
			"field":"status","value":"[\"active\",\"pending\"]","valueType":"List","priority":0
		}
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"status": "active"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("status 'active' IN list should be true, got ok=%v err=%v", ok, err)
	}
}

// ─── RegexpCondition via FromJSON ─────────────────────────────────────────────

func TestDefaultEngine_RegexpCondition_Match(t *testing.T) {
	raw := `{
		"id":"c1","name":"email regexp",
		"type": "condition",
		"sign": "REGEXP",
		"cacheable": false,
		"ignoreAbsence": false,
		"properties": {
			"field":"email","regexp":"^[a-z]+@example\\.com$","priority":0
		}
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"email": "alice@example.com"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("email should match regexp, got ok=%v err=%v", ok, err)
	}
}

// ─── CEL condition via FromJSON ───────────────────────────────────────────────

func TestDefaultEngine_CELCondition_True(t *testing.T) {
	raw := `{
		"id":"c1","name":"cel check",
		"type": "condition",
		"sign": "CEL",
		"cacheable": false,
		"ignoreAbsence": false,
		"properties": {
			"expression":"age >= 18","priority":0
		}
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{"age": int64(20)})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("CEL age>=18 at 20 should be true, got ok=%v err=%v", ok, err)
	}
}

// ─── EXISTS / ABSENT via FromJSON ─────────────────────────────────────────────

func TestDefaultEngine_ExistsCondition_FieldPresent(t *testing.T) {
	raw := `{
		"id":"c1","name":"exists check",
		"type": "condition",
		"sign": "EXISTS",
		"cacheable": false,
		"ignoreAbsence": false,
		"properties": {
			"field":"params.name","priority":0
		}
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{
		"params": map[string]interface{}{"name": "alice"},
	})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("field present → exists should be true, got ok=%v err=%v", ok, err)
	}
}

func TestDefaultEngine_AbsentCondition_FieldMissing(t *testing.T) {
	raw := `{
		"id":"c1","name":"absent check",
		"type": "condition",
		"sign": "ABSENT",
		"cacheable": false,
		"ignoreAbsence": false,
		"properties": {
			"field":"params.age","priority":0
		}
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()
	ctx := engine.NewRuleContext(map[string]interface{}{
		"params": map[string]interface{}{"name": "alice"},
	})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("field absent → absent condition should be true, got ok=%v err=%v", ok, err)
	}
}

// ─── NOT rule via FromJSON ────────────────────────────────────────────────────

func TestDefaultEngine_NOTRule_NegatesCondition(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "NOT",
		"children": [
			{
				"id":"c1","name":"n","type": "condition", "sign": "EQ", "cacheable": false, "ignoreAbsence": false,
				"properties": {"field":"role","value":"admin","valueType":"String","priority":0}
			}
		]
	}`
	rule := buildRule(t, "r1", raw)
	eng := engine.NewDefaultEngine()

	// NOT(role==admin) with role=user → true
	ctx := engine.NewRuleContext(map[string]interface{}{"role": "user"})
	ok, err := eng.Evaluate(rule, ctx)
	if err != nil || !ok {
		t.Fatalf("NOT(role==admin) with role=user should be true, got ok=%v err=%v", ok, err)
	}

	// NOT(role==admin) with role=admin → false
	ctx2 := engine.NewRuleContext(map[string]interface{}{"role": "admin"})
	ok2, err2 := eng.Evaluate(rule, ctx2)
	if err2 != nil || ok2 {
		t.Fatalf("NOT(role==admin) with role=admin should be false")
	}
}
