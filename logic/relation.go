package logic

import (
	"errors"
	"strconv"

	"github.com/franklee-labs/celero-go/core"
)

func resolve(n core.Node) (core.Relation, error) {
	if n.Type() == core.RelationNode {
		tmp, err := n.(core.Relation).Resolve()
		if err != nil {
			return nil, err
		}
		return tmp, nil
	} else if n.Type() == core.ConditionNode {
		return NewANDWithPath(core.NewPath(append(make([]core.Condition, 0, 1), n.(core.Condition)))), nil
	} else {
		return nil, errors.New("unsupported NodeKind[" + strconv.Itoa(int(n.Type())) + "]")
	}
}

func negate(n core.Node) (core.Node, error) {
	if n.Type() == core.RelationNode {
		tmp, err := n.(core.Relation).Negate()
		if err != nil {
			return nil, err
		}
		return tmp.Resolve()
	} else if n.Type() == core.ConditionNode {
		tmp, err := n.(core.Condition).Negate()
		if err != nil {
			return nil, err
		}
		return NewANDWithPath(core.NewPath(append(make([]core.Condition, 0, 1), tmp))), nil
	} else {
		return nil, errors.New("unsupported NodeKind[" + strconv.Itoa(int(n.Type())) + "]")
	}
}

// ------------ AND

type AND struct {
	core.BaseRelation
}

func NewAND() *AND {
	return &AND{
		*core.NewBaseRelation(),
	}
}

func NewANDWithPath(path *core.Path) *AND {
	a := NewAND()
	if path != nil {
		a.SetPathGroup(core.NewPathGroup(path))
	}
	return a
}

func (a *AND) RelKind() core.RelationKind {
	return core.RelationAnd
}

func (a *AND) Validate() core.Validation {
	if len(a.Children()) == 0 {
		return core.Validation{Valid: false, Message: "And node must have at least one node."}
	}
	return core.ValidationOK
}

func (a *AND) Resolve() (core.Relation, error) {
	if len(a.Children()) == 1 {
		tmp, err := resolve(a.Children()[0])
		if err != nil {
			return nil, err
		}
		return tmp, nil
	}
	var current core.Relation
	for i := 1; i < len(a.Children()); i++ {
		if current == nil {
			tmp, err := resolve(a.Children()[0])
			if err != nil {
				return nil, err
			}
			current = tmp
		}
		var next core.Relation
		n, err1 := resolve(a.Children()[i])
		if err1 != nil {
			return nil, err1
		}
		next = n
		var merged core.Relation
		if current.RelKind() == core.RelationAnd {
			p1 := current.PathGroup().Get(0)
			if next.RelKind() == core.RelationAnd {
				p2 := next.PathGroup().Get(0)
				conds := make([]core.Condition, 0, p1.Size()+p2.Size())
				conds = append(conds, p1.Conditions()...)
				conds = append(conds, p2.Conditions()...)
				merged = NewANDWithPath(core.NewPath(conds))
			} else {
				group := core.NewPathGroup()
				for _, p2 := range next.PathGroup().Paths() {
					conds := make([]core.Condition, 0, p1.Size()+p2.Size())
					conds = append(conds, p1.Conditions()...)
					conds = append(conds, p2.Conditions()...)
					group.AppendPath(core.NewPath(conds))
				}
				or := NewOR()
				or.SetPathGroup(group)
				merged = or
			}
		} else {
			paths := current.PathGroup().Paths()
			group := core.NewPathGroup()
			if next.RelKind() == core.RelationAnd {
				p2 := next.PathGroup().Paths()[0]
				for _, p := range paths {
					conds := make([]core.Condition, 0, p.Size()+p2.Size())
					conds = append(conds, p.Conditions()...)
					conds = append(conds, p2.Conditions()...)
					group.AppendPath(core.NewPath(conds))
				}
				or := NewOR()
				or.SetPathGroup(group)
				merged = or
			} else {
				ps := next.PathGroup().Paths()
				for _, p := range paths {
					for _, p2 := range ps {
						conds := make([]core.Condition, 0, p.Size()+p2.Size())
						conds = append(conds, p.Conditions()...)
						conds = append(conds, p2.Conditions()...)
						group.AppendPath(core.NewPath(conds))
					}
				}
				or := NewOR()
				or.SetPathGroup(group)
				merged = or
			}
		}
		current = merged
	}
	return current, nil
}

func (a *AND) Negate() (core.Relation, error) {
	or := NewOR()
	nodes := make([]core.Node, 0, len(a.Children()))
	for _, child := range a.Children() {
		tmp, err := negate(child)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, tmp)
	}
	or.SetChildren(nodes)
	return or, nil
}

// ------------ OR

type OR struct {
	core.BaseRelation
}

func NewOR() *OR {
	return &OR{
		*core.NewBaseRelation(),
	}
}

func (o *OR) RelKind() core.RelationKind {
	return core.RelationOr
}

func (o *OR) Validate() core.Validation {
	if len(o.Children()) < 2 {
		return core.Validation{Valid: false, Message: "Or node must have at least two nodes."}
	}
	return core.ValidationOK
}

func (o *OR) Resolve() (core.Relation, error) {
	var current core.Relation
	for i := 1; i < len(o.Children()); i++ {
		if current == nil {
			tmp, err := resolve(o.Children()[0])
			if err != nil {
				return nil, err
			}
			current = tmp
		}
		var next core.Relation
		n := o.Children()[i]
		var (
			tmp core.Relation
			err error
		)
		if n.Type() == core.RelationNode {
			tmp, err = n.(core.Relation).Resolve()
			if err != nil {
				return nil, err
			}
			next = tmp
		} else if n.Type() == core.ConditionNode {
			next = NewANDWithPath(core.NewPath(append(make([]core.Condition, 0, 1), n.(core.Condition))))
		} else {
			return nil, errors.New("unsupported NodeKind[" + strconv.Itoa(int(n.Type())) + "]")
		}
		group := core.NewPathGroup()
		for _, p := range current.PathGroup().Paths() {
			group.AppendPath(core.NewPath(p.Conditions()))
		}
		for _, p := range next.PathGroup().Paths() {
			group.AppendPath(core.NewPath(p.Conditions()))
		}
		or := NewOR()
		or.SetPathGroup(group)
		current = or
	}
	return current, nil
}

func (o *OR) Negate() (core.Relation, error) {
	and := NewAND()
	nodes := make([]core.Node, 0, len(o.Children()))
	for _, child := range o.Children() {
		tmp, err := negate(child)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, tmp)
	}
	and.SetChildren(nodes)
	return and, nil
}

// ------------ NOT

type NOT struct {
	core.BaseRelation
}

func NewNOT() *NOT {
	return &NOT{
		*core.NewBaseRelation(),
	}
}

func (n *NOT) RelKind() core.RelationKind {
	return core.RelationNot
}

func (n *NOT) Validate() core.Validation {
	if len(n.Children()) != 1 {
		return core.Validation{Valid: false, Message: "Not node must hold exactly one node."}
	}
	return core.ValidationOK
}

func (n *NOT) Resolve() (core.Relation, error) {
	node, err := negate(n.Children()[0])
	if err != nil {
		return nil, err
	}
	return resolve(node)
}

func (n *NOT) Negate() (core.Relation, error) {
	node := n.Children()[0]
	if node.Type() == core.RelationNode {
		return node.(core.Relation), nil
	}
	and := NewAND()
	and.SetChildren([]core.Node{node})
	return and, nil
}
