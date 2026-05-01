package engine

type Item struct {
	conditionID   string
	conditionName string
}

func createItem(conditionID, conditionName string) Item {
	return Item{conditionID: conditionID, conditionName: conditionName}
}

func (i *Item) ConditionID() string {
	return i.conditionID
}

func (i *Item) ConditionName() string {
	return i.conditionName
}

type Route struct {
	matched   []Item
	unmatched []Item
	absent    []Item
	skipped   []Item
}


func (r *Route) Matched() []Item {
	return r.matched
}

func (r *Route) Unmatched() []Item {
	return r.unmatched
}

func (r *Route) Absent() []Item {
	return r.absent
}

func (r *Route) Skipped() []Item {
	return r.skipped
}

type Report struct {
	routes []*Route
}

func (r *Report) appendRoute(route *Route) {
	r.routes = append(r.routes, route)
}

func (r *Report) Routes() []*Route {
	return r.routes
}
