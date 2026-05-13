package rules_test

import (
	"encoding/json"
	"testing"

	"github.com/franklee-labs/celero-go/core"
	"github.com/franklee-labs/celero-go/rules"
)

// ─── GetConditionCreator ─────────────────────────────────────────────────────

func TestGetConditionCreator_AllBuiltins(t *testing.T) {
	signs := []string{
		"EQ", "NEQ", "LT", "LTE", "GT", "GTE",
		"IN", "NIN", "REGEXP", "CEL", "EXISTS", "ABSENT",
		"INTERSECT", "DISJOINT",
	}
	for _, sign := range signs {
		creator, err := rules.GetConditionCreator("anyRule", sign)
		if err != nil {
			t.Errorf("GetConditionCreator(%q) unexpected error: %v", sign, err)
		}
		if creator == nil {
			t.Errorf("GetConditionCreator(%q) returned nil creator", sign)
		}
	}
}

func TestGetConditionCreator_CaseInsensitive(t *testing.T) {
	_, err := rules.GetConditionCreator("r1", "eq")
	if err != nil {
		t.Fatalf("GetConditionCreator should be case-insensitive, got error: %v", err)
	}
}

func TestGetConditionCreator_EmptySign(t *testing.T) {
	_, err := rules.GetConditionCreator("r1", "")
	if err == nil {
		t.Fatal("empty sign should return error")
	}
}

func TestGetConditionCreator_UnknownSign(t *testing.T) {
	_, err := rules.GetConditionCreator("r1", "UNKNOWN_SIGN_XYZ")
	if err == nil {
		t.Fatal("unknown sign should return error")
	}
}

// ─── RegisterRuleConditionCreator ────────────────────────────────────────────

func TestRegisterRuleConditionCreator_Success(t *testing.T) {
	creator := rules.ConditionCreateFunc(func(node rules.ConditionNode) (core.Condition, error) {
		return nil, nil
	})
	err := rules.RegisterRuleConditionCreator("rule-register-test", "CUSTOM_RULE_OP", creator)
	if err != nil {
		t.Fatalf("RegisterRuleConditionCreator error: %v", err)
	}
	got, err := rules.GetConditionCreator("rule-register-test", "CUSTOM_RULE_OP")
	if err != nil || got == nil {
		t.Fatalf("rule-specific creator not found after register: err=%v", err)
	}
}

func TestRegisterRuleConditionCreator_Duplicate(t *testing.T) {
	creator := rules.ConditionCreateFunc(func(node rules.ConditionNode) (core.Condition, error) {
		return nil, nil
	})
	_ = rules.RegisterRuleConditionCreator("rule-dup-test", "DUPOP", creator)
	err := rules.RegisterRuleConditionCreator("rule-dup-test", "DUPOP", creator)
	if err == nil {
		t.Fatal("duplicate rule-specific creator should return error")
	}
}

func TestRegisterRuleConditionCreator_EmptySign(t *testing.T) {
	creator := rules.ConditionCreateFunc(func(node rules.ConditionNode) (core.Condition, error) {
		return nil, nil
	})
	err := rules.RegisterRuleConditionCreator("r1", "", creator)
	if err == nil {
		t.Fatal("empty sign should return error")
	}
}

func TestRegisterRuleConditionCreator_InternalSignRejected(t *testing.T) {
	creator := rules.ConditionCreateFunc(func(node rules.ConditionNode) (core.Condition, error) {
		return nil, nil
	})
	for _, reserved := range []string{"AND", "OR", "NOT"} {
		err := rules.RegisterRuleConditionCreator("r1", reserved, creator)
		if err == nil {
			t.Fatalf("reserved sign %q should be rejected", reserved)
		}
	}
}

// Rule-specific creator takes precedence; unknown sign for other rules errors.
func TestGetConditionCreator_RuleSpecificOverridesGlobal(t *testing.T) {
	customCreator := rules.ConditionCreateFunc(func(node rules.ConditionNode) (core.Condition, error) {
		return nil, nil
	})
	ruleID := "rule-override-unique-123"
	sign := "MY_OVERRIDE_OP_456"
	if err := rules.RegisterRuleConditionCreator(ruleID, sign, customCreator); err != nil {
		t.Fatalf("register error: %v", err)
	}
	got, err := rules.GetConditionCreator(ruleID, sign)
	if err != nil || got == nil {
		t.Fatalf("rule-specific creator should be found: err=%v", err)
	}
	_, err = rules.GetConditionCreator("different-rule", sign)
	if err == nil {
		t.Fatalf("sign %q should not be available for different-rule", sign)
	}
}

// ─── StringValue / IntValue / BoolValue ──────────────────────────────────────

func TestStringValue_Present(t *testing.T) {
	props := map[string]interface{}{"key": "hello"}
	v, err := rules.StringValue(props, "key")
	if err != nil || v != "hello" {
		t.Fatalf("expected 'hello', got %q err=%v", v, err)
	}
}

func TestStringValue_Missing(t *testing.T) {
	_, err := rules.StringValue(map[string]interface{}{}, "key")
	if err == nil {
		t.Fatal("missing key should return error")
	}
}

func TestStringValue_WrongType(t *testing.T) {
	props := map[string]interface{}{"key": 42}
	_, err := rules.StringValue(props, "key")
	if err == nil {
		t.Fatal("non-string value should return error")
	}
}

func TestIntValue_Present(t *testing.T) {
	props := map[string]interface{}{"priority": float64(5)}
	v, err := rules.IntValue(props, "priority")
	if err != nil || v != 5 {
		t.Fatalf("expected 5, got %d err=%v", v, err)
	}
}

func TestIntValue_Missing(t *testing.T) {
	_, err := rules.IntValue(map[string]interface{}{}, "priority")
	if err == nil {
		t.Fatal("missing key should return error")
	}
}

func TestIntValue_WrongType(t *testing.T) {
	props := map[string]interface{}{"priority": "high"}
	_, err := rules.IntValue(props, "priority")
	if err == nil {
		t.Fatal("non-float64 value should return error")
	}
}

func TestBoolValue_Present(t *testing.T) {
	props := map[string]interface{}{"flag": true}
	v, err := rules.BoolValue(props, "flag")
	if err != nil || !v {
		t.Fatalf("expected true, got %v err=%v", v, err)
	}
}

func TestBoolValue_Missing(t *testing.T) {
	_, err := rules.BoolValue(map[string]interface{}{}, "flag")
	if err == nil {
		t.Fatal("missing key should return error")
	}
}

func TestBoolValue_WrongType(t *testing.T) {
	props := map[string]interface{}{"flag": "yes"}
	_, err := rules.BoolValue(props, "flag")
	if err == nil {
		t.Fatal("non-bool value should return error")
	}
}

// ─── BuildConditionConfig ────────────────────────────────────────────────────

func TestBuildConditionConfig_Valid(t *testing.T) {
	node := rules.ConditionNode{}
	node.Props = map[string]interface{}{
		"id":       "c1",
		"name":     "my condition",
		"priority": float64(0),
	}
	cfg, err := rules.BuildConditionConfig(node)
	if err != nil {
		t.Fatalf("BuildConditionConfig error: %v", err)
	}
	if cfg.ID() != "c1" || cfg.Name() != "my condition" {
		t.Fatalf("unexpected config values: id=%q name=%q", cfg.ID(), cfg.Name())
	}
}

func TestBuildConditionConfig_MissingID(t *testing.T) {
	node := rules.ConditionNode{}
	node.Props = map[string]interface{}{
		"name":     "my condition",
		"priority": float64(0),
	}
	_, err := rules.BuildConditionConfig(node)
	if err == nil {
		t.Fatal("missing 'id' should return error")
	}
}

func TestBuildConditionConfig_MissingName(t *testing.T) {
	node := rules.ConditionNode{}
	node.Props = map[string]interface{}{
		"id":       "c1",
		"priority": float64(0),
	}
	_, err := rules.BuildConditionConfig(node)
	if err == nil {
		t.Fatal("missing 'name' should return error")
	}
}

func TestBuildConditionConfig_MissingPriority(t *testing.T) {
	node := rules.ConditionNode{}
	node.Props = map[string]interface{}{
		"id":   "c1",
		"name": "my condition",
	}
	_, err := rules.BuildConditionConfig(node)
	if err == nil {
		t.Fatal("missing 'priority' should return error")
	}
}

func TestBuildConditionConfig_CacheableAndIgnoreAbsence(t *testing.T) {
	node := rules.ConditionNode{Cacheable: true, IgnoreAbsence: true}
	node.Props = map[string]interface{}{
		"id":       "c1",
		"name":     "n",
		"priority": float64(0),
	}
	cfg, err := rules.BuildConditionConfig(node)
	if err != nil {
		t.Fatalf("BuildConditionConfig error: %v", err)
	}
	if !cfg.Cacheable() || !cfg.IgnoreAbsence() {
		t.Fatal("cacheable/ignoreAbsence not propagated correctly")
	}
}

// ─── ConditionNode.Transform ─────────────────────────────────────────────────

func makeConditionNode(sign string, props map[string]interface{}) rules.ConditionNode {
	n := rules.ConditionNode{}
	n.Sign = sign
	n.Type = "condition"
	n.Props = props
	return n
}

func baseProps(field, value, valueType string) map[string]interface{} {
	return map[string]interface{}{
		"id":        "c1",
		"name":      "test-cond",
		"priority":  float64(0),
		"field":     field,
		"value":     value,
		"valueType": valueType,
	}
}

func TestConditionNodeTransform_EQ(t *testing.T) {
	node := makeConditionNode("EQ", baseProps("age", "18", "Number"))
	meta := rules.CreateRuleMeta("rule1", "Rule 1")
	n, err := node.Transform(meta)
	if err != nil {
		t.Fatalf("Transform EQ error: %v", err)
	}
	if n == nil {
		t.Fatal("Transform returned nil node")
	}
	if n.Type() != core.ConditionNode {
		t.Fatal("expected ConditionNode type")
	}
}

func TestConditionNodeTransform_UnknownSign(t *testing.T) {
	node := makeConditionNode("UNKNOWN_XYZ", map[string]interface{}{})
	meta := rules.CreateRuleMeta("rule1", "")
	_, err := node.Transform(meta)
	if err == nil {
		t.Fatal("unknown sign should return error")
	}
}

func TestConditionNodeTransform_InvalidProperties(t *testing.T) {
	// Missing required fields → BuildConditionConfig fails
	node := makeConditionNode("EQ", map[string]interface{}{"id": "c1"})
	meta := rules.CreateRuleMeta("rule1", "")
	_, err := node.Transform(meta)
	if err == nil {
		t.Fatal("missing required properties should return error")
	}
}

// ─── RelationNode ────────────────────────────────────────────────────────────

func TestCreateRelationNode_AND(t *testing.T) {
	n := rules.CreateRelationNode("AND")
	if n.Kind() != "relation" {
		t.Fatalf("expected kind 'relation', got %q", n.Kind())
	}
	if n.Sign != "AND" {
		t.Fatalf("expected sign AND, got %q", n.Sign)
	}
}

func TestCreateRelationNode_UnsupportedSign(t *testing.T) {
	meta := rules.CreateRuleMeta("r1", "")
	n := rules.CreateRelationNode("INVALID_SIGN")
	_, err := n.Transform(meta)
	if err == nil {
		t.Fatal("unsupported relation sign should return error")
	}
}

func TestRelationNode_UnmarshalJSON_SingleConditionChild(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "AND",
		"children": [
			{
				"type": "condition",
				"sign": "EQ",
				"cacheable": false,
				"ignoreAbsence": false,
				"properties": {
					"id": "c1", "name": "age check",
					"field": "age", "value": "18", "valueType": "Number", "priority": 0
				}
			}
		]
	}`
	var rn rules.RelationNode
	if err := json.Unmarshal([]byte(raw), &rn); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if len(rn.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(rn.Children))
	}
	if rn.Children[0].Kind() != "condition" {
		t.Fatalf("child kind should be 'condition', got %q", rn.Children[0].Kind())
	}
}

func TestRelationNode_UnmarshalJSON_NestedRelation(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "OR",
		"children": [
			{
				"type": "relation",
				"sign": "AND",
				"children": [
					{
						"type": "condition",
						"sign": "EQ",
						"properties": {
							"id": "c1", "name": "n", "field": "a", "value": "1",
							"valueType": "String", "priority": 0
						}
					}
				]
			},
			{
				"type": "condition",
				"sign": "EQ",
				"properties": {
					"id": "c2", "name": "n2", "field": "b", "value": "2",
					"valueType": "String", "priority": 0
				}
			}
		]
	}`
	var rn rules.RelationNode
	if err := json.Unmarshal([]byte(raw), &rn); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if len(rn.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(rn.Children))
	}
	if rn.Children[0].Kind() != "relation" {
		t.Fatalf("first child should be relation, got %q", rn.Children[0].Kind())
	}
	if rn.Children[1].Kind() != "condition" {
		t.Fatalf("second child should be condition, got %q", rn.Children[1].Kind())
	}
}

func TestRelationNode_UnmarshalJSON_UnknownChildType(t *testing.T) {
	raw := `{
		"type": "relation",
		"sign": "AND",
		"children": [{"type": "unknown_type"}]
	}`
	var rn rules.RelationNode
	err := json.Unmarshal([]byte(raw), &rn)
	if err == nil {
		t.Fatal("unknown child type should return error")
	}
}
