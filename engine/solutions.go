package engine

import "github.com/franklee-labs/celero-go/core"

type Solutions struct {
	solutions []Solution
}

func createSolutions(group *core.PathGroup) Solutions {
	solutions := make([]Solution, 0, group.Size())
	for i := 0; i < group.Size(); i++ {
		p := group.Get(i)
		conditions := make([]Condition, 0, p.Size())
		for j := 0; j < p.Size(); j++ {
			cond := p.Get(j)
			conditions = append(conditions, Condition{id: cond.ID(), name: cond.Name()})
		}
		solutions = append(solutions, Solution{conditions: conditions})
	}
	return Solutions{solutions: solutions}
}

type Solution struct {
	conditions []Condition
}

func (s Solution) Conditions() []Condition {
	return s.conditions
}

type Condition struct {
	id   string
	name string
}

func (c Condition) ID() string {
	return c.id
}

func (c Condition) Name() string {
	return c.name
}
