package rules

import (
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

func (n RelationNode) Transform(m RuleMeta) (core.Relation, error) {
	node, err := n.createRelatioon()
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

func (n RelationNode) createRelatioon() (core.Relation, error) {
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

func (n ConditionNode) Transform(m RuleMeta) (core.Node, error) {
	return nil, nil
}
