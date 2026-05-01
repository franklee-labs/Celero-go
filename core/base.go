package core

import (
	"github.com/google/uuid"
)

type BaseNode struct {
	id       string
	name     string
	nodeKind NodeKind
	nodes    []Node
	ruleID   string
}

func (b *BaseNode) SetID(id string) *BaseNode {
	b.id = id
	return b
}
func (b *BaseNode) ID() string { return b.id }
func (b *BaseNode) SetName(name string) *BaseNode {
	b.name = name
	return b
}
func (b *BaseNode) Name() string         { return b.name }
func (b *BaseNode) Type() NodeKind       { return b.nodeKind }
func (b *BaseNode) Children() []Node     { return b.nodes }
func (b *BaseNode) SetChildren(c []Node) { b.nodes = c }
func (b *BaseNode) RuleID() string       { return b.ruleID }
func (b *BaseNode) SetRuleID(id string)  { b.ruleID = id }

type BaseRelation struct {
	BaseNode
	pg *PathGroup
}

func NewBaseRelation() *BaseRelation {
	id := uuid.NewString()
	return &BaseRelation{
		BaseNode: BaseNode{id: id, name: id, nodeKind: RelationNode},
	}
}

func (r *BaseRelation) PathGroup() *PathGroup      { return r.pg }
func (r *BaseRelation) SetPathGroup(pg *PathGroup) { r.pg = pg }

type BaseCondition struct {
	BaseNode
	internalId    string
	priority      int
	cacheable     bool
	ignoreAbsence bool
}

func NewBaseCondition(cfg *ConditionConfig) BaseCondition {
	return BaseCondition{
		BaseNode:      BaseNode{id: cfg.ID(), name: cfg.Name(), nodeKind: ConditionNode},
		internalId:    uuid.New().String(),
		priority:      cfg.Priority(),
		cacheable:     cfg.Cacheable(),
		ignoreAbsence: cfg.IgnoreAbsence(),
	}
}

func (c *BaseCondition) InternalID() string  { return c.internalId }
func (c *BaseCondition) Cacheable() bool     { return c.cacheable }
func (c *BaseCondition) Priority() int       { return c.priority }
func (c *BaseCondition) IgnoreAbsence() bool { return c.ignoreAbsence }

func ValidateAll(n Node) Validation {
	if n.Type() == ConditionNode {
		if v := n.(Condition).Validate(); !v.Valid {
			return v
		}
		return ValidationOK
	} else if n.Type() == RelationNode {
		if v := n.(Relation).Validate(); !v.Valid {
			return v
		}
		for _, child := range n.(Relation).Children() {
			if v := ValidateAll(child); !v.Valid {
				return v
			}
		}
		return ValidationOK
	} else {
		return ValidationOK
	}
}
