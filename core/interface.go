package core

type RuleContext interface {
	GetParams() map[string]interface{}
	GetAttributes() map[string]interface{}
	GetAttribute(key string) (interface{}, bool)
	SetAttribute(key string, value interface{})
	IsReportEnabled() bool
}

type EvalContext interface {
	Params() map[string]interface{}
	IsConditionResultCacheEnabled() bool
	IsAbsenceEnabled() bool
	IsReportEnabled() bool
	BuildEvalParams(p map[string]interface{}) error
	EvalParams() map[string]interface{}
	RuleContext() RuleContext
	SetConditionEvalResult(conditionID string, result EvalResult)
	GetConditionEvalResult(conditionID string) (EvalResult, bool)
}

type Node interface {
	ID() string
	Name() string
	Type() NodeKind
	RuleID() string
	SetRuleID(string)
}

type Relation interface {
	Node
	Children() []Node
	SetChildren([]Node)
	Validate() Validation
	RelKind() RelationKind
	PathGroup() *PathGroup
	SetPathGroup(*PathGroup)
	Resolve() (Relation, error)
	Negate() (Relation, error)
}

type Condition interface {
	Node

	// impletemented by BaseCondition
	InternalID() string
	IgnoreAbsence() bool
	Cacheable() bool
	Priority() int

	// unimplemented methods
	Validate() Validation
	Expression() string
	Evaluate(ctx EvalContext) (bool, bool, error)
	Negate() (Condition, error)
	Compile() error
	BeforeEvaluate(ctx EvalContext) error
}
