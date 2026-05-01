package core

import (
	"math"
)

const PriorityDefault int = 0
const PriorityHighest int = math.MinInt32
const PriorityLowest int = math.MaxInt32

type ConditionConfig struct {
	id            string
	name          string
	priority      int
	cacheable     bool
	ignoreAbsence bool
}

func NewConditionConfig(id, name string) *ConditionConfig {
	return &ConditionConfig{
		id:            id,
		name:          name,
		priority:      PriorityDefault,
		cacheable:     false,
		ignoreAbsence: false,
	}
}

func (c *ConditionConfig) WithPriority(p int) *ConditionConfig {
	if p > PriorityLowest {
		p = PriorityLowest
	}
	if p < PriorityHighest {
		p = PriorityHighest
	}
	c.priority = p
	return c
}

func (c *ConditionConfig) WithCacheable(b bool) *ConditionConfig {
	c.cacheable = b
	return c
}

func (c *ConditionConfig) WithIgnoreAbsence(b bool) *ConditionConfig {
	c.ignoreAbsence = b
	return c
}

func (c *ConditionConfig) ID() string {
	return c.id
}

func (c *ConditionConfig) Name() string {
	return c.name
}

func (c *ConditionConfig) Priority() int {
	return c.priority
}

func (c *ConditionConfig) Cacheable() bool {
	return c.cacheable
}

func (c *ConditionConfig) IgnoreAbsence() bool {
	return c.ignoreAbsence
}
