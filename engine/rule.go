package engine

import (
	"fmt"
	"sync/atomic"

	"github.com/franklee-labs/celero-go/core"
)

// Rule is the compiled internal representation of a business rule.
// It holds the expanded PathGroup produced during Build and is wrapped by CeleroRule before being handed to users.
type Rule struct {
	id          string
	name        string
	description string
	cacheable   atomic.Bool
	root        core.Relation
	pathGroup   *core.PathGroup
}

func (r *Rule) PathGroup() *core.PathGroup {
	return r.pathGroup
}

func (r *Rule) ID() string {
	return r.id
}

func (r *Rule) Name() string {
	return r.name
}

func (r *Rule) Description() string {
	return r.description
}

func (r *Rule) Cacheable() bool {
	return r.cacheable.Load()
}

func (r *Rule) SetCacheable(c bool) {
	r.cacheable.Store(c)
}

// Build validates the rule tree and expands it into a flat PathGroup.
// It must be called once before the rule can be evaluated; RuleBuilder.Build does this automatically.
func (r *Rule) Build() error {
	if v := core.ValidateAll(r.root); !v.Valid {
		return fmt.Errorf("invalid rule %s", v.Message)
	}
	relation, err := r.root.Resolve()
	if err != nil {
		return err
	}
	r.pathGroup = relation.PathGroup()
	r.pathGroup.Sort()
	return nil
}

func createRule(id, name, description string, cacheable bool, root core.Relation) *Rule {
	r := &Rule{
		id:          id,
		name:        name,
		description: description,
		root:        root,
	}
	r.SetCacheable(cacheable)
	return r
}
