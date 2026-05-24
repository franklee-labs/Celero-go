// Demonstrates ConditionListener and RuleListener with explicit priority ordering,
// and how listeners share state via RuleContext.SetAttribute / GetAttribute.
//
// Context attribute flow:
//
//	"metrics.total" / "metrics.matched"  written by metricsConditionListener (order=1)
//	                                      read    by debugConditionListener   (order=10)
//	"failed.conditions"                  written by alertConditionListener   (order=20)
//	                                      read    by auditRuleListener        (order=1)
//	"matched.rules"                      written by rewardRuleListener       (order=5)
//	                                      read    by main() after evaluation
//
// Listeners are registered out of order intentionally — the engine sorts them by Order().
package main

import (
	"fmt"

	"github.com/franklee-labs/celero-go/engine"
	"github.com/franklee-labs/celero-go/examples/shared"
)

const (
	attrMetricsTotal    = "metrics.total"
	attrMetricsMatched  = "metrics.matched"
	attrFailedConditions = "failed.conditions"
	attrMatchedRules    = "matched.rules"
)

// ── Condition listeners ───────────────────────────────────────────────────────

// order=1: increments running counters in context — fires first
type metricsConditionListener struct{}

func (l *metricsConditionListener) OnResult(event *engine.ConditionEvent) {
	ctx := event.Context()
	total := getInt(ctx, attrMetricsTotal) + 1
	matched := getInt(ctx, attrMetricsMatched)
	if event.Matched() {
		matched++
	}
	ctx.SetAttribute(attrMetricsTotal, total)
	ctx.SetAttribute(attrMetricsMatched, matched)
	fmt.Printf("  [CONDITION][order=1 ][METRICS] setAttribute: total=%d  matched=%d\n", total, matched)
}

func (l *metricsConditionListener) Order() int { return 1 }

// order=10: reads the counters written by metricsConditionListener — fires second
type debugConditionListener struct{}

func (l *debugConditionListener) OnResult(event *engine.ConditionEvent) {
	ctx := event.Context()
	fmt.Printf("  [CONDITION][order=10][DEBUG] getAttribute: total=%d matched=%d  |  rule=%s  condition=%s  result=%v\n",
		getInt(ctx, attrMetricsTotal),
		getInt(ctx, attrMetricsMatched),
		event.RuleName(), event.ConditionName(), event.Matched())
}

func (l *debugConditionListener) Order() int { return 10 }

// order=20: appends failed condition names to context list — fires last
type alertConditionListener struct{}

func (l *alertConditionListener) OnResult(event *engine.ConditionEvent) {
	if !event.Matched() {
		ctx := event.Context()
		failed := getStringSlice(ctx, attrFailedConditions)
		failed = append(failed, event.ConditionName())
		ctx.SetAttribute(attrFailedConditions, failed)
		fmt.Printf("  [CONDITION][order=20][ALERT] setAttribute: failed.conditions=%v\n", failed)
	}
}

func (l *alertConditionListener) Order() int { return 20 }

// ── Rule listeners ────────────────────────────────────────────────────────────

// order=1: reads and clears "failed.conditions" accumulated this rule cycle — fires first
type auditRuleListener struct{}

func (l *auditRuleListener) OnResult(event *engine.RuleEvent) {
	ctx := event.Context()
	failed := getStringSlice(ctx, attrFailedConditions)
	pass := "FAIL"
	if event.Matched() {
		pass = "PASS"
	}
	fmt.Printf("  [RULE][order=1 ][AUDIT] getAttribute: failed.conditions=%v  |  rule=%s  result=%s\n",
		failed, event.RuleName(), pass)
	ctx.SetAttribute(attrFailedConditions, []string{})
}

func (l *auditRuleListener) Order() int { return 1 }

// order=5: writes matched rule IDs to context for post-evaluation summary — fires second
type rewardRuleListener struct{}

func (l *rewardRuleListener) OnResult(event *engine.RuleEvent) {
	if event.Matched() {
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

	eng := engine.NewDefaultEngine()

	// Register listeners out of order intentionally — engine sorts by Order()
	eng.AddConditionListener(&alertConditionListener{})   // order=20
	eng.AddConditionListener(&debugConditionListener{})   // order=10
	eng.AddConditionListener(&metricsConditionListener{}) // order=1

	eng.AddRuleListener(&rewardRuleListener{}) // order=5
	eng.AddRuleListener(&auditRuleListener{})  // order=1

	user := map[string]interface{}{
		"name":           "Alice",
		"totalSpend":     int64(800),
		"memberLevel":    "gold",
		"registeredDays": int64(365),
		"orderCount":     int64(20),
		"status":         "active",
		"age":            int64(25),
		"verified":       true,
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
