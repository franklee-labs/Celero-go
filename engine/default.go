package engine

import (
	"slices"
	"sort"

	"github.com/franklee-labs/celero-go/core"
)

func buildContext(ctx *RuleContext, enableConditionResultCache, enableMissingState bool) (*EvalContext, error) {
	builder := newBuilder(ctx).
		SetConditionResultCacheable(enableConditionResultCache).
		SetEnableMissingState(enableMissingState).
		EnableReport(ctx.IsReportEnabled())
	context, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return context, nil
}

func execute(condition core.Condition, context core.EvalContext) (rtn core.EvalResult, err error) {
	condition.BeforeEvaluate(context)
	result, absent, er := condition.Evaluate(context)
	defer func() {
		if context.IsConditionResultCacheEnabled() && condition.Cacheable() {
			context.SetConditionEvalResult(condition.InternalID(), rtn)
		}
	}()
	if absent {
		if !condition.IgnoreAbsence() && context.IsAbsenceEnabled() {
			rtn = core.EvalResultIndeterminate
			err = nil
			return
		}
		rtn = core.EvalResultFalse
		err = nil
		return
	}
	if er != nil {
		rtn = core.EvalResultUnknown
		err = er
		return
	}
	if result {
		rtn = core.EvalResultTrue
		err = nil
		return
	}
	rtn = core.EvalResultFalse
	err = nil
	return
}

func createRoute(matchedIdx []int, unmatchedIdx []int, skippedAt int, conditions []core.Condition) *Route {
	route := &Route{}
	matched := make([]Item, 0, len(matchedIdx))
	unmatched := make([]Item, 0, len(unmatchedIdx))
	skipped := make([]Item, 0)
	for i, cond := range conditions {
		if i >= skippedAt {
			skipped = append(skipped, createItem(cond.ID(), cond.Name()))
		} else {
			if slices.Contains(matchedIdx, i) {
				matched = append(matched, createItem(cond.ID(), cond.Name()))
			} else if slices.Contains(unmatchedIdx, i) {
				unmatched = append(unmatched, createItem(cond.ID(), cond.Name()))
			}
		}
	}
	route.matched = matched
	route.unmatched = unmatched
	route.absent = skipped
	return route
}

type DefaultEngine struct {
	enableMissingState bool
	conditionListeners []ConditionListener
	ruleListeners      []RuleListener
}

func NewDefaultEngine() *DefaultEngine {
	return &DefaultEngine{
		enableMissingState: false,
		conditionListeners: make([]ConditionListener, 0),
		ruleListeners:      make([]RuleListener, 0),
	}
}

func (e *DefaultEngine) AddConditionListener(listener ConditionListener) {
	e.conditionListeners = append(e.conditionListeners, listener)
	sort.Slice(e.conditionListeners, func(i, j int) bool {
		return e.conditionListeners[i].Order() < e.conditionListeners[j].Order()
	})
}

func (e *DefaultEngine) AddRuleListener(listener RuleListener) {
	e.ruleListeners = append(e.ruleListeners, listener)
	sort.Slice(e.ruleListeners, func(i, j int) bool {
		return e.ruleListeners[i].Order() < e.ruleListeners[j].Order()
	})
}

func (e *DefaultEngine) Evalutes(rules []*CeleroRule, ctx *RuleContext) {
	for _, rule := range rules {
		result, err := e.Evaluate(rule, ctx)
		if err == nil {
			event := newRuleEvent(rule.rule().ID(), rule.rule().Name(), result, ctx)
			e.callRuleListeners(event)
		}
	}
}

func (e *DefaultEngine) Evaluate(rule *CeleroRule, ctx *RuleContext) (bool, error) {
	context, err := buildContext(ctx, rule.IsCacheable(), false)
	if err != nil {
		return false, err
	}
	for _, path := range rule.rule().PathGroup().Paths() {
		rt, err := e.execute(rule, path, context)
		if err != nil {
			return false, err
		}
		if rt {
			return true, nil
		}
	}
	return false, nil
}

func (e *DefaultEngine) execute(rule *CeleroRule, path *core.Path, context *EvalContext) (bool, error) {
	matchedIdx := make([]int, 0)
	unmatchedIdx := make([]int, 0)
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
				return false, err
			}
			result = &rtn
			event := newConditionEvent(rule.rule().ID(), rule.rule().Name(), cond.ID(), cond.Name(), rtn.IsTrue(), context.ruleContext)
			e.callConditionListeners(event)
			if !rtn.IsTrue() {
				if context.IsReportEnabled() {
					unmatchedIdx = append(unmatchedIdx, i)
					skippedAt = i + 1
					route := createRoute(matchedIdx, unmatchedIdx, skippedAt, path.Conditions())
					context.ruleContext.appendRoute(rule, route)
				}
				return false, nil
			} else {
				matchedIdx = append(matchedIdx, i)
			}
		} else if !result.IsTrue() {
			if context.IsReportEnabled() {
				unmatchedIdx = append(unmatchedIdx, i)
				skippedAt = i + 1
				route := createRoute(matchedIdx, unmatchedIdx, skippedAt, path.Conditions())
				context.ruleContext.appendRoute(rule, route)
			}
			return false, nil
		} else {
			matchedIdx = append(matchedIdx, i)
		}
	}
	if context.IsReportEnabled() {
		route := createRoute(matchedIdx, unmatchedIdx, skippedAt, path.Conditions())
		context.ruleContext.appendRoute(rule, route)
	}
	return true, nil
}

func (e *DefaultEngine) callConditionListeners(event *ConditionEvent) {
	for _, listener := range e.conditionListeners {
		listener.OnResult(event)
	}
}

func (e *DefaultEngine) callRuleListeners(event *RuleEvent) {
	for _, listener := range e.ruleListeners {
		listener.OnResult(event)
	}
}
