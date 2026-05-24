package engine

import "github.com/franklee-labs/celero-go/core"

// ConditionEvent carries the result of a single condition evaluation in DefaultEngine.
type ConditionEvent struct {
	ruleID        string
	ruleName      string
	conditionID   string
	conditionName string
	matched       bool
	err           error
	context       *RuleContext
}

func newConditionEvent(ruleID, ruleName, conditionID, conditionName string, matched bool, ctx *RuleContext) *ConditionEvent {
	return &ConditionEvent{
		ruleID:        ruleID,
		ruleName:      ruleName,
		conditionID:   conditionID,
		conditionName: conditionName,
		matched:       matched,
		context:       ctx,
	}
}

// RuleID returns the ID of the rule being evaluated.
func (e *ConditionEvent) RuleID() string {
	return e.ruleID
}

// RuleName returns the name of the rule being evaluated.
func (e *ConditionEvent) RuleName() string {
	return e.ruleName
}

// ConditionID returns the ID of the condition that was just evaluated.
func (e *ConditionEvent) ConditionID() string {
	return e.conditionID
}

// ConditionName returns the name of the condition that was just evaluated.
func (e *ConditionEvent) ConditionName() string {
	return e.conditionName
}

// Matched reports whether the condition evaluated to true.
func (e *ConditionEvent) Matched() bool {
	return e.matched
}

// Context returns the shared RuleContext for this evaluation, allowing listeners to read and write attributes.
func (e *ConditionEvent) Context() *RuleContext {
	return e.context
}

// ConditionListener is implemented by types that want to observe condition evaluation results in DefaultEngine.
// Order controls execution sequence: lower values run first.
type ConditionListener interface {
	OnResult(event *ConditionEvent)
	Order() int
}

// RuleEvent carries the overall result of a single rule evaluation in DefaultEngine.
// It is fired by Evalutes after each rule; Evaluate does not fire it.
type RuleEvent struct {
	ruleID   string
	ruleName string
	matched  bool
	context  *RuleContext
}

func newRuleEvent(ruleID, ruleName string, matched bool, ctx *RuleContext) *RuleEvent {
	return &RuleEvent{
		ruleID:   ruleID,
		ruleName: ruleName,
		matched:  matched,
		context:  ctx,
	}
}

// RuleID returns the ID of the evaluated rule.
func (e *RuleEvent) RuleID() string {
	return e.ruleID
}

// RuleName returns the name of the evaluated rule.
func (e *RuleEvent) RuleName() string {
	return e.ruleName
}

// Matched reports whether the rule evaluated to true.
func (e *RuleEvent) Matched() bool {
	return e.matched
}

// Context returns the shared RuleContext, allowing listeners to read and write attributes.
func (e *RuleEvent) Context() *RuleContext {
	return e.context
}

// RuleListener is implemented by types that want to observe rule evaluation results in DefaultEngine.
// Order controls execution sequence: lower values run first.
type RuleListener interface {
	OnResult(event *RuleEvent)
	Order() int
}

// AdvancedConditionEvent carries the three-valued EvalResult of a condition evaluation in AdvancedEngine.
type AdvancedConditionEvent struct {
	ruleID        string
	ruleName      string
	conditionID   string
	conditionName string
	result        core.EvalResult
	context       *RuleContext
}

func newAdvancedConditionEvent(ruleID, ruleName, conditionID, conditionName string, result core.EvalResult, ctx *RuleContext) *AdvancedConditionEvent {
	return &AdvancedConditionEvent{
		ruleID:        ruleID,
		ruleName:      ruleName,
		conditionID:   conditionID,
		conditionName: conditionName,
		result:        result,
		context:       ctx,
	}
}

// RuleID returns the ID of the rule being evaluated.
func (e *AdvancedConditionEvent) RuleID() string {
	return e.ruleID
}

// RuleName returns the name of the rule being evaluated.
func (e *AdvancedConditionEvent) RuleName() string {
	return e.ruleName
}

// ConditionID returns the ID of the condition that was just evaluated.
func (e *AdvancedConditionEvent) ConditionID() string {
	return e.conditionID
}

// ConditionName returns the name of the condition that was just evaluated.
func (e *AdvancedConditionEvent) ConditionName() string {
	return e.conditionName
}

// Result returns the three-valued evaluation result: True, False, or Indeterminate.
func (e *AdvancedConditionEvent) Result() core.EvalResult {
	return e.result
}

// Context returns the shared RuleContext, allowing listeners to read and write attributes.
func (e *AdvancedConditionEvent) Context() *RuleContext {
	return e.context
}

// AdvancedConditionListener is implemented by types that want to observe condition results in AdvancedEngine.
// Events carry a full EvalResult, enabling listeners to react to Indeterminate (missing-field) cases.
// Order controls execution sequence: lower values run first.
type AdvancedConditionListener interface {
	OnResult(event *AdvancedConditionEvent)
	Order() int
}

// AdvancedRuleEvent carries the three-valued EvalResult of a rule evaluation in AdvancedEngine.
// It is fired by Evalutes after each rule; Evaluate does not fire it.
type AdvancedRuleEvent struct {
	ruleID   string
	ruleName string
	matched  core.EvalResult
	context  *RuleContext
}

func newAdvancedRuleEvent(ruleID, ruleName string, matched core.EvalResult, ctx *RuleContext) *AdvancedRuleEvent {
	return &AdvancedRuleEvent{
		ruleID:   ruleID,
		ruleName: ruleName,
		matched:  matched,
		context:  ctx,
	}
}

// RuleID returns the ID of the evaluated rule.
func (e *AdvancedRuleEvent) RuleID() string {
	return e.ruleID
}

// RuleName returns the name of the evaluated rule.
func (e *AdvancedRuleEvent) RuleName() string {
	return e.ruleName
}

// Matched returns the three-valued rule result: True, False, or Indeterminate.
func (e *AdvancedRuleEvent) Matched() core.EvalResult {
	return e.matched
}

// Context returns the shared RuleContext, allowing listeners to read and write attributes.
func (e *AdvancedRuleEvent) Context() *RuleContext {
	return e.context
}

// AdvancedRuleListener is implemented by types that want to observe rule evaluation results in AdvancedEngine.
// Order controls execution sequence: lower values run first.
type AdvancedRuleListener interface {
	OnResult(event *AdvancedRuleEvent)
	Order() int
}
