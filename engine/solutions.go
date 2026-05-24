package engine

import "github.com/franklee-labs/celero-go/core"

// Solutions holds the expanded path structure of a compiled rule.
// Use CeleroRule.Solutions() to retrieve it and inspect the paths before or after evaluation.
type Solutions struct {
	solutions []Solution
}

// Count returns the total number of expanded paths in the rule.
func (s Solutions) Count() int {
	return len(s.solutions)
}

// SolutionAt returns the Solution at the given index, along with a boolean indicating whether the index was valid.
func (s Solutions) SolutionAt(index int) (Solution, bool) {
	if index < 0 || index >= len(s.solutions) {
		return Solution{}, false
	}
	return s.solutions[index], true
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

// Solution represents a single expanded path: an ordered list of conditions that must all pass for the rule to match.
type Solution struct {
	conditions []Condition
}

// Conditions returns all conditions in this path in evaluation order.
func (s Solution) Conditions() []Condition {
	return s.conditions
}

// Count returns the number of conditions in this path.
func (s Solution) Count() int {
	return len(s.conditions)
}

// ConditionAt returns the Condition at the given index, along with a boolean indicating whether the index was valid.
func (s Solution) ConditionAt(index int) (Condition, bool) {
	if index < 0 || index >= len(s.conditions) {
		return Condition{}, false
	}
	return s.conditions[index], true
}

// Condition is a lightweight snapshot of a condition node within a Solution, holding its ID and name.
type Condition struct {
	id   string
	name string
}

// ID returns the condition's unique identifier.
func (c Condition) ID() string {
	return c.id
}

// Name returns the condition's human-readable name.
func (c Condition) Name() string {
	return c.name
}
