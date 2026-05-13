package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/franklee-labs/celero-go/core"
	"github.com/franklee-labs/celero-go/rules"
)

type RuleBuilder struct {
	id          string
	name        string
	description string
	cacheable   bool
	root        *rules.RelationNode
}

func (b *RuleBuilder) ID(id string) *RuleBuilder {
	b.id = id
	return b
}

func (b *RuleBuilder) Name(name string) *RuleBuilder {
	b.name = name
	return b
}

func (b *RuleBuilder) Description(description string) *RuleBuilder {
	b.description = description
	return b
}

func (b *RuleBuilder) Cacheable(cacheable bool) *RuleBuilder {
	b.cacheable = cacheable
	return b
}

func (b *RuleBuilder) Root(root *rules.RelationNode) *RuleBuilder {
	b.root = root
	return b
}

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

func NewRuleBuilder() *RuleBuilder {
	return &RuleBuilder{}
}

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
