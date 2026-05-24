package engine

// Item represents a single condition entry within a Route, identified by its ID and name.
type Item struct {
	conditionID   string
	conditionName string
}

func createItem(conditionID, conditionName string) Item {
	return Item{conditionID: conditionID, conditionName: conditionName}
}

// ConditionID returns the condition's unique identifier.
func (i *Item) ConditionID() string {
	return i.conditionID
}

// ConditionName returns the condition's human-readable name.
func (i *Item) ConditionName() string {
	return i.conditionName
}

// Route records the evaluation outcome of a single path attempt.
// A rule's Report may contain multiple Routes when the rule expands into multiple paths.
type Route struct {
	matched   []Item
	unmatched []Item
	absent    []Item
	skipped   []Item
}

// Matched returns the conditions that evaluated to true on this path.
func (r *Route) Matched() []Item {
	return r.matched
}

// Unmatched returns the condition that caused this path to fail (at most one, due to short-circuit).
func (r *Route) Unmatched() []Item {
	return r.unmatched
}

// Absent returns the conditions that were skipped because an earlier condition failed (short-circuit tail).
func (r *Route) Absent() []Item {
	return r.absent
}

// Skipped returns conditions not evaluated on this path. Reserved for future use; currently always empty.
func (r *Route) Skipped() []Item {
	return r.skipped
}

// Report aggregates all Route records produced during a single rule evaluation.
// Retrieve it from RuleContext.Reports() after calling Evaluate or Evalutes with reports enabled.
type Report struct {
	routes []*Route
}

func (r *Report) appendRoute(route *Route) {
	r.routes = append(r.routes, route)
}

// Routes returns all path evaluation records for this rule.
func (r *Report) Routes() []*Route {
	return r.routes
}
