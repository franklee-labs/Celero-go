package core

import "math"

type NodeKind int

const (
	_             NodeKind = iota // Skip 0
	RelationNode                  // 1
	ConditionNode                 // 2
)

// --------------------------------

type RelationKind int

const (
	_           RelationKind = iota // Skip 0
	RelationAnd                     // 1
	RelationOr                      // 2
	RelationNot                     // 3
)

// --------------------------------

var (
	defaultPriority = 0
	lowestPriority  = math.MaxInt32
	highestPriority = math.MinInt32
)

// --------------------------------

type State int

const (
	Unknown State = iota // Skip 0
	True
	False
	Indeterminate
)

type EvalResult struct {
	state State
}

func (r EvalResult) IsTrue() bool {
	return r.state == True
}

func (r EvalResult) IsFalse() bool {
	return r.state == False
}

func (r EvalResult) IsIndeterminate() bool {
	return r.state == Indeterminate
}

func (r EvalResult) IsUnknown() bool {
	return r.state == Unknown
}

var EvalResultUnknown = EvalResult{state: Unknown}
var EvalResultTrue = EvalResult{state: True}
var EvalResultFalse = EvalResult{state: False}
var EvalResultIndeterminate = EvalResult{state: Indeterminate}

// --------------------------------

type Validation struct {
	Valid   bool
	Message string
}

var ValidationOK = Validation{Valid: true}

func Invalid(message string) Validation {
	return Validation{Valid: false, Message: message}
}

// --------------------------------

type ValueType int

const (
	ValueTypeInvalid ValueType = iota
	ValueTypeString
	ValueTypeNumber
	ValueTypeBool
	ValueTypeList
	ValueTypeExpression
)

func ValueTypeFromString(s string) ValueType {
	switch s {
	case "String":
		return ValueTypeString
	case "Number":
		return ValueTypeNumber
	case "Boolean":
		return ValueTypeBool
	case "List":
		return ValueTypeList
	case "Expression":
		return ValueTypeExpression
	default:
		return ValueTypeInvalid
	}
}
