package engine

import (
	"sort"

	"github.com/franklee-labs/celero-go/core"
)

type AdvancedEngine struct {
	enableMissingState bool
	conditionListeners []AdvancedConditionListener
	ruleListeners      []AdvancedRuleListener
}

func NewAdvancedEngine() *AdvancedEngine {
	return &AdvancedEngine{
		enableMissingState: true,
		conditionListeners: make([]AdvancedConditionListener, 0),
		ruleListeners:      make([]AdvancedRuleListener, 0),
	}
}

func (e *AdvancedEngine) AddConditionListener(listener AdvancedConditionListener) {
	e.conditionListeners = append(e.conditionListeners, listener)
	sort.Slice(e.conditionListeners, func(i, j int) bool {
		return e.conditionListeners[i].Order() < e.conditionListeners[j].Order()
	})
}

func (e *AdvancedEngine) AddRuleListener(listener AdvancedRuleListener) {
	e.ruleListeners = append(e.ruleListeners, listener)
	sort.Slice(e.ruleListeners, func(i, j int) bool {
		return e.ruleListeners[i].Order() < e.ruleListeners[j].Order()
	})
}

func (e *AdvancedEngine) Evalutes(rules []*CeleroRule, ctx *RuleContext) {
	for _, rule := range rules {
		result, err := e.Evaluate(rule, ctx)
		if err == nil {
			event := newAdvancedRuleEvent(rule.rule().ID(), rule.rule().Name(), result, ctx)
			e.callRuleListeners(event)
		}
	}
}

func (e *AdvancedEngine) Evaluate(rule *CeleroRule, ctx *RuleContext) (core.EvalResult, error) {
	context, err := buildContext(ctx, rule.IsCacheable(), true)
	if err != nil {
		return core.EvalResultUnknown, err
	}
	miss := false
	for _, path := range rule.rule().PathGroup().Paths() {
		rt, err := e.execute(rule, path, context)
		if err != nil {
			return core.EvalResultUnknown, err
		}
		if rt.IsTrue() {
			return rt, nil
		}
		if rt.IsIndeterminate() {
			miss = true
		}
	}
	if miss {
		return core.EvalResultIndeterminate, nil
	}
	return core.EvalResultFalse, nil
}

func (e *AdvancedEngine) execute(rule *CeleroRule, path *core.Path, context *EvalContext) (core.EvalResult, error) {
	absent := false
	matchedIdx := make([]int, 0)
	unmatchedIdx := make([]int, 0)
	absentIdx := make([]int, 0)
	skippedAt := path.Size()
	for i := 0; i < path.Size(); i++ {
		var result *core.EvalResult = nil
		var hitCache bool = false
		cond := path.Get(i)
		if context.IsConditionResultCacheEnabled() {
			rt, exists := context.GetConditionEvalResult(cond.InternalID())
			hitCache = exists
			if exists {
				result = &rt
			}
		}
		if !hitCache {
			rtn, err := execute(cond, context)
			if err != nil {
				return core.EvalResultUnknown, err
			}
			result = &rtn
			event := newAdvancedConditionEvent(rule.rule().ID(), rule.rule().Name(), cond.ID(), cond.Name(), *result, context.ruleContext)
			e.callConditionListeners(event)
			if result.IsFalse() {
				unmatchedIdx = append(unmatchedIdx, i)
				skippedAt = i + 1
				if context.IsReportEnabled() {
					route := createRoute(matchedIdx, unmatchedIdx, skippedAt, path.Conditions())
					context.ruleContext.appendRoute(rule, route)
				}
				return core.EvalResultFalse, nil
			} else if result.IsIndeterminate() {
				absentIdx = append(absentIdx, i)
				absent = true
			} else {
				matchedIdx = append(matchedIdx, i)
			}

		} else if result.IsFalse() {
			unmatchedIdx = append(unmatchedIdx, i)
			skippedAt = i + 1
			if context.IsReportEnabled() {
				route := createRoute(matchedIdx, unmatchedIdx, skippedAt, path.Conditions())
				context.ruleContext.appendRoute(rule, route)
			}
			return core.EvalResultFalse, nil
		} else if result.IsIndeterminate() {
			absentIdx = append(absentIdx, i)
			absent = true
		} else {
			matchedIdx = append(matchedIdx, i)
		}
	}
	if context.IsReportEnabled() {
		route := createRoute(matchedIdx, unmatchedIdx, skippedAt, path.Conditions())
		context.ruleContext.appendRoute(rule, route)
	}
	if absent {
		return core.EvalResultIndeterminate, nil
	}
	return core.EvalResultTrue, nil
}

func (e *AdvancedEngine) callConditionListeners(event *AdvancedConditionEvent) {
	for _, listener := range e.conditionListeners {
		listener.OnResult(event)
	}
}

func (e *AdvancedEngine) callRuleListeners(event *AdvancedRuleEvent) {
	for _, listener := range e.ruleListeners {
		listener.OnResult(event)
	}
}
