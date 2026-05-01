package engine

type RuleContext struct {
	params        map[string]interface{}
	attributes    map[string]interface{}
	enableReports bool
	reports       map[*CeleroRule]*Report
}

func NewRuleContext(params map[string]interface{}) *RuleContext {
	return &RuleContext{
		params:        params,
		attributes:    make(map[string]interface{}),
		enableReports: false,
		reports:       make(map[*CeleroRule]*Report),
	}
}

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

func (rc *RuleContext) Reports() map[*CeleroRule]*Report {
	return rc.reports
}

func (rc *RuleContext) GetAttributes() map[string]interface{} {
	return rc.attributes
}

func (rc *RuleContext) SetAttribute(key string, value interface{}) {
	rc.attributes[key] = value
}

func (rc *RuleContext) GetAttribute(key string) (interface{}, bool) {
	value, exists := rc.attributes[key]
	return value, exists
}

func (rc *RuleContext) EnableReports() {
	rc.enableReports = true
}

func (rc *RuleContext) IsReportEnabled() bool {
	return rc.enableReports
}

type CeleroRule struct {
	r         *Rule
	solutions Solutions
}

func CreateCeleroRule(rule *Rule) *CeleroRule {
	return &CeleroRule{r: rule, solutions: createSolutions(rule.PathGroup())}
}

func (cr *CeleroRule) rule() *Rule {
	return cr.r
}

func (cr *CeleroRule) Solutions() Solutions {
	return cr.solutions
}

func (cr *CeleroRule) ID() string {
	return cr.r.ID()
}

func (cr *CeleroRule) IsCacheable() bool {
	return cr.r.Cacheable()
}

func (cr *CeleroRule) SetCacheable(c bool) {
	cr.r.SetCacheable(c)
}
