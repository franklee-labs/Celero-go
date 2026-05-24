package engine

// RuleContext carries the input parameters and shared state for a single evaluation request.
// It is created once per request and may be reused across multiple rules in a batch (Evalutes).
// Attributes stored via SetAttribute are visible to all listeners and remain accessible after evaluation.
type RuleContext struct {
	params        map[string]interface{}
	attributes    map[string]interface{}
	enableReports bool
	reports       map[*CeleroRule]*Report
}

// NewRuleContext creates a RuleContext with the given parameters.
// params is the data map that conditions read from (e.g. {"age": int64(25), "status": "active"}).
func NewRuleContext(params map[string]interface{}) *RuleContext {
	return &RuleContext{
		params:        params,
		attributes:    make(map[string]interface{}),
		enableReports: false,
		reports:       make(map[*CeleroRule]*Report),
	}
}

// GetParams returns the input parameter map passed at construction.
func (rc *RuleContext) GetParams() map[string]interface{} {
	return rc.params
}

func (rc *RuleContext) appendRoute(rule *CeleroRule, route *Route) {
	if report, has := rc.reports[rule]; has {
		report.appendRoute(route)
	} else {
		report := &Report{routes: make([]*Route, 0)}
		report.appendRoute(route)
		rc.reports[rule] = report
	}
}

// Reports returns the per-rule evaluation reports. Only populated when EnableReports has been called.
func (rc *RuleContext) Reports() map[*CeleroRule]*Report {
	return rc.reports
}

// GetAttributes returns all listener-written attributes as a map.
func (rc *RuleContext) GetAttributes() map[string]interface{} {
	return rc.attributes
}

// SetAttribute stores an arbitrary value under key, visible to all subsequent listeners in this evaluation.
func (rc *RuleContext) SetAttribute(key string, value interface{}) {
	rc.attributes[key] = value
}

// GetAttribute retrieves a previously stored attribute. Returns (value, true) if found, (nil, false) otherwise.
func (rc *RuleContext) GetAttribute(key string) (interface{}, bool) {
	value, exists := rc.attributes[key]
	return value, exists
}

// EnableReports turns on per-rule evaluation reporting.
// When enabled, each Evaluate call records matched, unmatched, and skipped conditions per path.
// Retrieve results via Reports() after evaluation.
func (rc *RuleContext) EnableReports() {
	rc.enableReports = true
}

// IsReportEnabled reports whether evaluation reporting is active.
func (rc *RuleContext) IsReportEnabled() bool {
	return rc.enableReports
}

// CeleroRule is the user-facing handle for a compiled rule.
// Obtain one through RuleBuilder.Build or FromJSON(...).Build.
type CeleroRule struct {
	r         *Rule
	solutions Solutions
}

// CreateCeleroRule wraps an internal Rule into a CeleroRule. Intended for advanced or test use; prefer RuleBuilder.
func CreateCeleroRule(rule *Rule) *CeleroRule {
	return &CeleroRule{r: rule, solutions: createSolutions(rule.PathGroup())}
}

func (cr *CeleroRule) rule() *Rule {
	return cr.r
}

// Solutions returns the expanded path structure of the rule, useful for inspecting what paths exist before evaluation.
func (cr *CeleroRule) Solutions() Solutions {
	return cr.solutions
}

// ID returns the rule identifier set during construction.
func (cr *CeleroRule) ID() string {
	return cr.r.ID()
}

// Name returns the human-readable rule name.
func (cr *CeleroRule) Name() string {
	return cr.r.Name()
}

// Description returns the rule description.
func (cr *CeleroRule) Description() string {
	return cr.r.Description()
}

// IsCacheable reports whether cross-path condition result caching is enabled for this rule.
func (cr *CeleroRule) IsCacheable() bool {
	return cr.r.Cacheable()
}

// SetCacheable toggles cross-path condition result caching at runtime.
func (cr *CeleroRule) SetCacheable(c bool) {
	cr.r.SetCacheable(c)
}
