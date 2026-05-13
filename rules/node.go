package rules

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/franklee-labs/celero-go/core"
	"github.com/franklee-labs/celero-go/logic"
)

type ConfigNode interface {
	Kind() string
	Transform(meta RuleMeta) (core.Node, error)
}

type RuleMeta struct {
	id   string
	name string
}

func CreateRuleMeta(id, name string) RuleMeta {
	return RuleMeta{id, name}
}

func (m RuleMeta) ID() string {
	return m.id
}

func (m RuleMeta) Name() string {
	return m.name
}

type BaseNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Sign string `json:"sign"`
}

type RelationNode struct {
	BaseNode
	Children []ConfigNode `json:"children"`
}

func CreateRelationNode(sign string, children ...ConfigNode) RelationNode {
	return RelationNode{
		BaseNode{
			ID:   "",
			Name: "",
			Type: "relation",
			Sign: strings.ToUpper(sign),
		},
		append(make([]ConfigNode, 0, len(children)), children...),
	}
}

func (n RelationNode) Kind() string {
	return n.Type
}

// UnmarshalJSON handles nested RelationNode/ConditionNode children from JSON.
func (n *RelationNode) UnmarshalJSON(data []byte) error {
	type alias struct {
		BaseNode
		Children []json.RawMessage `json:"children"`
	}
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	n.BaseNode = tmp.BaseNode
	n.Children = make([]ConfigNode, 0, len(tmp.Children))
	for _, raw := range tmp.Children {
		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &base); err != nil {
			return err
		}
		switch strings.ToLower(base.Type) {
		case "relation":
			var r RelationNode
			if err := json.Unmarshal(raw, &r); err != nil {
				return err
			}
			n.Children = append(n.Children, r)
		case "condition":
			var c ConditionNode
			if err := json.Unmarshal(raw, &c); err != nil {
				return err
			}
			n.Children = append(n.Children, c)
		default:
			return fmt.Errorf("unknown node type %q", base.Type)
		}
	}
	return nil
}

// Transform converts RelationNode into a core.Relation with fully-built children.
func (n RelationNode) Transform(m RuleMeta) (core.Node, error) {
	node, err := n.createRelation()
	if err != nil {
		return nil, err
	}
	node.SetRuleID(m.ID())
	tmp := make([]core.Node, 0, len(n.Children))
	for _, child := range n.Children {
		c, err := child.Transform(m)
		if err != nil {
			return nil, err
		}
		tmp = append(tmp, c)
	}
	node.SetChildren(tmp)
	return node, nil
}

func (n RelationNode) createRelation() (core.Relation, error) {
	if strings.EqualFold("and", n.Sign) {
		and := logic.NewAND()
		and.SetID(n.ID).SetName(n.Name)
		return and, nil
	} else if strings.EqualFold("or", n.Sign) {
		or := logic.NewOR()
		or.SetID(n.ID).SetName(n.Name)
		return or, nil
	} else if strings.EqualFold("not", n.Sign) {
		not := logic.NewNOT()
		not.SetID(n.ID).SetName(n.Name)
		return not, nil
	}
	return nil, fmt.Errorf("unsupported relation type. supported[AND, OR, NOT]")
}

type ConditionNode struct {
	BaseNode
	Cacheable     bool                   `json:"cacheable"`
	IgnoreAbsence bool                   `json:"ignoreAbsence"`
	Props         map[string]interface{} `json:"properties"`
}

func (n ConditionNode) Kind() string {
	return n.Type
}

func (n ConditionNode) IsCacheable() bool {
	return n.Cacheable
}

func (n ConditionNode) SetCacheable(b bool) {
	n.Cacheable = b
}

func (n ConditionNode) IsIgnoreAbsence() bool {
	return n.IgnoreAbsence
}

func (n ConditionNode) SetIgnoreAbsence(b bool) {
	n.IgnoreAbsence = b
}

func (n ConditionNode) Properties() map[string]interface{} {
	return n.Props
}

func (n ConditionNode) SetProperties(p map[string]interface{}) {
	n.Props = p
}

// Transform builds a compiled core.Condition using the registered ConditionCreator.
func (n ConditionNode) Transform(m RuleMeta) (core.Node, error) {
	creator, err := GetConditionCreator(m.ID(), n.Sign)
	if err != nil {
		return nil, fmt.Errorf("no creator for condition sign %q: %w", n.Sign, err)
	}
	cond, err := creator.Create(n)
	if err != nil {
		return nil, fmt.Errorf("create condition %q: %w", n.Sign, err)
	}
	cond.SetRuleID(m.ID())
	if v := cond.Validate(); !v.Valid {
		return nil, fmt.Errorf("invalid condition [%s]: %s", n.Sign, v.Message)
	}
	if err := cond.Compile(); err != nil {
		return nil, fmt.Errorf("compile condition [%s]: %w", n.Sign, err)
	}
	return cond, nil
}
