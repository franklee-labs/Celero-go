package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/franklee-labs/celero-go/core"
	"github.com/franklee-labs/celero-go/rules"
)

// RuleBuilder constructs a CeleroRule using a fluent API.
// Use NewRuleBuilder for programmatic construction or FromJSON to start from a JSON rule tree.
type RuleBuilder struct {
	id          string
	name        string
	description string
	cacheable   bool
	root        *rules.RelationNode
}

// ID sets the unique identifier for the rule. Required.
func (b *RuleBuilder) ID(id string) *RuleBuilder {
	b.id = id
	return b
}

// Name sets the human-readable name of the rule.
func (b *RuleBuilder) Name(name string) *RuleBuilder {
	b.name = name
	return b
}

// Description sets a free-text description of the rule's intent.
func (b *RuleBuilder) Description(description string) *RuleBuilder {
	b.description = description
	return b
}

// Cacheable enables or disables cross-path condition result caching for this rule.
// Individual conditions must also set cacheable=true to participate in caching.
func (b *RuleBuilder) Cacheable(cacheable bool) *RuleBuilder {
	b.cacheable = cacheable
	return b
}

// Root sets the root relation node of the rule logic tree.
func (b *RuleBuilder) Root(root *rules.RelationNode) *RuleBuilder {
	b.root = root
	return b
}

// Build validates the rule tree, compiles all conditions, and returns a ready-to-evaluate CeleroRule.
// Returns an error if the rule is structurally invalid or any condition fails to compile.
func (b *RuleBuilder) Build() (*CeleroRule, error) {
	if b.root == nil {
		return nil, fmt.Errorf("root must not be nil")
	}
	if b.id == "" {
		return nil, fmt.Errorf("rule id must not be nil")
	}
	ruleMeta := rules.CreateRuleMeta(b.id, b.name)
	rawNode, err := b.root.Transform(ruleMeta)
	if err != nil {
		return nil, err
	}
	node, ok := rawNode.(core.Relation)
	if !ok {
		return nil, fmt.Errorf("failed to transform root node")
	}
	rule := createRule(b.id, b.name, b.description, b.cacheable, node)
	err = rule.Build()
	if err != nil {
		return nil, err
	}
	return CreateCeleroRule(rule), nil
}

// NewRuleBuilder returns an empty RuleBuilder for programmatic rule construction.
func NewRuleBuilder() *RuleBuilder {
	return &RuleBuilder{}
}

// FromJSON parses a JSON rule tree and returns a RuleBuilder pre-populated with the root node.
// id is the rule identifier; jsonRules is the JSON of the root relation or condition node.
// Call Name, Description, Cacheable, and Build on the returned builder to complete construction.
//
// JSON format: the root node must have a "type" field of "relation" or "condition".
// Condition properties (id, name, priority, field, value, etc.) are nested under "properties".
// The cacheable and ignoreAbsence flags are top-level fields on the condition node, not inside "properties".
func FromJSON(id, jsonRules string) (*RuleBuilder, error) {
	b := NewRuleBuilder()
	b.ID(id)
	var m map[string]interface{}
	msg := json.RawMessage(jsonRules)
	err := json.Unmarshal(msg, &m)
	if err != nil {
		return nil, err
	}
	if t, has := m["type"]; has {
		if tp, ok := t.(string); ok {
			if strings.ToLower(tp) == "relation" {
				var r rules.RelationNode
				err := json.Unmarshal(msg, &r)
				if err != nil {
					return nil, err
				}
				b.Root(&r)
				return b, nil
			} else if strings.ToLower(tp) == "condition" {
				var c rules.ConditionNode
				err := json.Unmarshal(msg, &c)
				if err != nil {
					return nil, err
				}
				b.Root(wrap(c))
				return b, nil
			} else {
				return nil, fmt.Errorf("invalid type")
			}
		}
		return nil, fmt.Errorf("invalid type")
	}
	return nil, fmt.Errorf("invalid type")

}

func wrap(n rules.ConditionNode) *rules.RelationNode {
	node := rules.CreateRelationNode("AND", n)
	return &node
}
