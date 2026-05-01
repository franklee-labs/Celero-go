package engine

import "github.com/franklee-labs/celero-go/core"

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

func (e *ConditionEvent) RuleID() string {
	return e.ruleID
}

func (e *ConditionEvent) RuleName() string {
	return e.ruleName
}

func (e *ConditionEvent) ConditionID() string {
	return e.conditionID
}

func (e *ConditionEvent) ConditionName() string {
	return e.conditionName
}

func (e *ConditionEvent) Matched() bool {
	return e.matched
}

func (e *ConditionEvent) Context() *RuleContext {
	return e.context
}

type ConditionListener interface {
	OnResult(event *ConditionEvent)
	Order() int
}

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

func (e *RuleEvent) RuleID() string {
	return e.ruleID
}

func (e *RuleEvent) RuleName() string {
	return e.ruleName
}

func (e *RuleEvent) Matched() bool {
	return e.matched
}

func (e *RuleEvent) Context() *RuleContext {
	return e.context
}

type RuleListener interface {
	OnResult(event *RuleEvent)
	Order() int
}

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

func (e *AdvancedConditionEvent) RuleID() string {
	return e.ruleID
}

func (e *AdvancedConditionEvent) RuleName() string {
	return e.ruleName
}

func (e *AdvancedConditionEvent) ConditionID() string {
	return e.conditionID
}

func (e *AdvancedConditionEvent) ConditionName() string {
	return e.conditionName
}

func (e *AdvancedConditionEvent) Result() core.EvalResult {
	return e.result
}

func (e *AdvancedConditionEvent) Context() *RuleContext {
	return e.context
}

type AdvancedConditionListener interface {
	OnResult(event *AdvancedConditionEvent)
	Order() int
}

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

func (e *AdvancedRuleEvent) RuleID() string {
	return e.ruleID
}

func (e *AdvancedRuleEvent) RuleName() string {
	return e.ruleName
}

func (e *AdvancedRuleEvent) Matched() core.EvalResult {
	return e.matched
}

func (e *AdvancedRuleEvent) Context() *RuleContext {
	return e.context
}

type AdvancedRuleListener interface {
	OnResult(event *AdvancedRuleEvent)
	Order() int
}
