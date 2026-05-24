// Demonstrates AdvancedConditionListener and AdvancedRuleListener with priority ordering,
// and how listeners share state via RuleContext.SetAttribute / GetAttribute.
//
// Key difference from simple_engine_listeners: condition events carry EvalResult
// (TRUE / FALSE / INDETERMINATE) instead of a plain boolean, allowing listeners to
// react specifically to missing-field cases.
//
// Context attribute flow:
//
//	"metrics.total" / "metrics.true" / "metrics.false" / "metrics.indeterminate"
//	                                   written by metricsConditionListener (order=1)
//	                                   read    by debugConditionListener   (order=10)
//	"indeterminate.conditions"         written by indeterminateConditionListener (order=20)
//	                                   read    by auditRuleListener        (order=1)
//	"matched.rules"                    written by rewardRuleListener       (order=5)
//	                                   read    by main() after evaluation
//
// Listeners are registered out of order intentionally — engine sorts by Order().
package main

import (
	"fmt"

	"github.com/franklee-labs/celero-go/core"
	"github.com/franklee-labs/celero-go/engine"
	"github.com/franklee-labs/celero-go/examples/shared"
)

const (
	attrMetricsTotal           = "metrics.total"
	attrMetricsTrue            = "metrics.true"
	attrMetricsFalse           = "metrics.false"
	attrMetricsIndeterminate   = "metrics.indeterminate"
	attrIndeterminateConditions = "indeterminate.conditions"
	attrMatchedRules           = "matched.rules"
)

// ── Condition listeners ───────────────────────────────────────────────────────

// order=1: increments per-state counters in context — fires first
type metricsConditionListener struct{}

func (l *metricsConditionListener) OnResult(event *engine.AdvancedConditionEvent) {
	ctx := event.Context()
	r := event.Result()
	total := getInt(ctx, attrMetricsTotal) + 1
	trueC := getInt(ctx, attrMetricsTrue)
	falseC := getInt(ctx, attrMetricsFalse)
	indetC := getInt(ctx, attrMetricsIndeterminate)
	if r.IsTrue() {
		trueC++
	} else if r.IsFalse() {
		falseC++
	} else if r.IsIndeterminate() {
		indetC++
	}
	ctx.SetAttribute(attrMetricsTotal, total)
	ctx.SetAttribute(attrMetricsTrue, trueC)
	ctx.SetAttribute(attrMetricsFalse, falseC)
	ctx.SetAttribute(attrMetricsIndeterminate, indetC)
	fmt.Printf("  [CONDITION][order=1 ][METRICS] setAttribute: total=%d  true=%d  false=%d  indeterminate=%d\n",
		total, trueC, falseC, indetC)
}

func (l *metricsConditionListener) Order() int { return 1 }

// order=10: reads counters written by metricsConditionListener — fires second
type debugConditionListener struct{}

func (l *debugConditionListener) OnResult(event *engine.AdvancedConditionEvent) {
	ctx := event.Context()
	fmt.Printf("  [CONDITION][order=10][DEBUG] getAttribute: true=%d false=%d indeterminate=%d  |  rule=%s  condition=%s  result=%s\n",
		getInt(ctx, attrMetricsTrue),
		getInt(ctx, attrMetricsFalse),
		getInt(ctx, attrMetricsIndeterminate),
		event.RuleName(), event.ConditionName(), labelShort(event.Result()))
}

func (l *debugConditionListener) Order() int { return 10 }

// order=20: appends INDETERMINATE condition names to context list — fires last
type indeterminateConditionListener struct{}

func (l *indeterminateConditionListener) OnResult(event *engine.AdvancedConditionEvent) {
	if event.Result().IsIndeterminate() && event.ConditionID() != "" {
		ctx := event.Context()
		list := getStringSlice(ctx, attrIndeterminateConditions)
		list = append(list, event.ConditionName())
		ctx.SetAttribute(attrIndeterminateConditions, list)
		fmt.Printf("  [CONDITION][order=20][INDET ] setAttribute: indeterminate.conditions=%v\n", list)
	}
}

func (l *indeterminateConditionListener) Order() int { return 20 }

// ── Rule listeners ────────────────────────────────────────────────────────────

// order=1: reads and clears "indeterminate.conditions" accumulated this rule cycle — fires first
type auditRuleListener struct{}

func (l *auditRuleListener) OnResult(event *engine.AdvancedRuleEvent) {
	ctx := event.Context()
	indet := getStringSlice(ctx, attrIndeterminateConditions)
	fmt.Printf("  [RULE][order=1 ][AUDIT] getAttribute: indeterminate.conditions=%v  |  rule=%s  result=%s\n",
		indet, event.RuleName(), labelShort(event.Matched()))
	ctx.SetAttribute(attrIndeterminateConditions, []string{})
}

func (l *auditRuleListener) Order() int { return 1 }

// order=5: writes matched rule IDs to context for post-evaluation summary — fires second
type rewardRuleListener struct{}

func (l *rewardRuleListener) OnResult(event *engine.AdvancedRuleEvent) {
	if event.Matched().IsTrue() {
		ctx := event.Context()
		matched := getStringSlice(ctx, attrMatchedRules)
		matched = append(matched, event.RuleID())
		ctx.SetAttribute(attrMatchedRules, matched)
		fmt.Printf("  [RULE][order=5 ][REWARD] setAttribute: matched.rules=%v\n", matched)
	}
}

func (l *rewardRuleListener) Order() int { return 5 }

// ── Setup & run ───────────────────────────────────────────────────────────────

func main() {
	rules, err := shared.LoadCouponRules()
	if err != nil {
		panic(err)
	}

	eng := engine.NewAdvancedEngine()

	// Register listeners out of order intentionally — engine sorts by Order()
	eng.AddConditionListener(&indeterminateConditionListener{}) // order=20
	eng.AddConditionListener(&debugConditionListener{})         // order=10
	eng.AddConditionListener(&metricsConditionListener{})       // order=1

	eng.AddRuleListener(&rewardRuleListener{}) // order=5
	eng.AddRuleListener(&auditRuleListener{})  // order=1

	// Grace: missing "age" and "verified" — vip-access will be INDETERMINATE
	user := map[string]interface{}{
		"name":           "Grace",
		"totalSpend":     int64(100),
		"memberLevel":    "normal",
		"registeredDays": int64(60),
		"orderCount":     int64(3),
		"status":         "active",
		"banned":         false,
	}

	fmt.Printf("Evaluating user: %s\n", user["name"])
	fmt.Println(repeat("─", 60))

	ctx := engine.NewRuleContext(user)
	eng.Evalutes(rules, ctx)

	fmt.Println(repeat("─", 60))
	matchedRules := getStringSlice(ctx, attrMatchedRules)
	fmt.Printf("[SUMMARY] getAttribute: matched.rules=%v\n", matchedRules)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func labelShort(r core.EvalResult) string {
	if r.IsTrue() {
		return "TRUE"
	}
	if r.IsFalse() {
		return "FALSE"
	}
	return "INDETERMINATE"
}

func getInt(ctx *engine.RuleContext, key string) int {
	val, ok := ctx.GetAttribute(key)
	if !ok {
		return 0
	}
	if i, ok := val.(int); ok {
		return i
	}
	return 0
}

func getStringSlice(ctx *engine.RuleContext, key string) []string {
	val, ok := ctx.GetAttribute(key)
	if !ok {
		return []string{}
	}
	if s, ok := val.([]string); ok {
		return s
	}
	return []string{}
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
