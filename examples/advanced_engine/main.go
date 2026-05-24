// Demonstrates AdvancedEngine which evaluates rules to a three-state result:
//
//	TRUE          — all conditions matched
//	FALSE         — at least one condition did not match
//	INDETERMINATE — no condition was false, but at least one required field was absent
//
// Compared to DefaultEngine (boolean), the advanced engine does not treat a
// missing field as an automatic FALSE — it defers the decision to the caller.
//
// Rule file: examples/shared/data/coupon-rules.json
package main

import (
	"fmt"

	"github.com/franklee-labs/celero-go/core"
	"github.com/franklee-labs/celero-go/engine"
	"github.com/franklee-labs/celero-go/examples/shared"
)

func main() {
	rules, err := shared.LoadCouponRules()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded rules: %d\n", len(rules))
	for _, r := range rules {
		fmt.Printf("  [%s] %s — %s\n", r.ID(), r.Name(), r.Description())
	}
	fmt.Println()

	runSingleRule(rules)
	runMultiRules(rules)
}

func runSingleRule(rules []*engine.CeleroRule) {
	eng := engine.NewAdvancedEngine()

	fmt.Println("=== Single rule evaluation ===")
	var vipRule *engine.CeleroRule
	for _, r := range rules {
		if r.ID() == "vip-access" {
			vipRule = r
			break
		}
	}

	// missing "age" and "verified" — OR branch cannot be resolved
	incompleteUser := map[string]interface{}{"status": "active", "banned": false}
	result, err := eng.Evaluate(vipRule, engine.NewRuleContext(incompleteUser))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rule [%s]  result: %s\n\n", vipRule.Name(), label(result))
}

func runMultiRules(rules []*engine.CeleroRule) {
	eng := engine.NewAdvancedEngine()

	users := []map[string]interface{}{
		// expected: TRUE  for high-value-user + vip-access
		{"name": "Alice", "totalSpend": int64(800), "memberLevel": "gold", "registeredDays": int64(365), "orderCount": int64(20), "status": "active", "age": int64(25), "verified": true, "banned": false},
		// expected: TRUE  for new-user-welcome + vip-access
		{"name": "Bob", "totalSpend": int64(0), "memberLevel": "normal", "registeredDays": int64(7), "orderCount": int64(0), "status": "active", "age": int64(20), "verified": false, "banned": false},
		// expected: INDETERMINATE for high-value-user (memberLevel absent), TRUE for vip-access
		{"name": "Frank", "totalSpend": int64(800), "registeredDays": int64(180), "orderCount": int64(10), "status": "active", "age": int64(28), "verified": true, "banned": false},
		// expected: FALSE for high-value-user (spend too low), INDETERMINATE for vip-access (age + verified absent)
		{"name": "Grace", "totalSpend": int64(100), "memberLevel": "normal", "registeredDays": int64(60), "orderCount": int64(3), "status": "active", "banned": false},
		// expected: FALSE for all rules
		{"name": "Henry", "totalSpend": int64(200), "memberLevel": "bronze", "registeredDays": int64(90), "orderCount": int64(8), "status": "inactive", "age": int64(35), "verified": false, "banned": true},
	}

	for _, user := range users {
		fmt.Printf("┌── User: %s\n", user["name"])

		ctx := engine.NewRuleContext(user)
		ctx.EnableReports()

		for _, r := range rules {
			result, err := eng.Evaluate(r, ctx)
			if err != nil {
				panic(err)
			}
			report := ctx.Reports()[r]
			fmt.Printf("│  %s [%s] %s\n", label(result), r.ID(), r.Name())
			if result.IsTrue() {
				printConditions("matched  ", report)
			} else if result.IsFalse() {
				printConditions("unmatched", report)
			}
		}

		fmt.Println("└───────────────────────────────────────")
		fmt.Println()
	}
}

func label(result core.EvalResult) string {
	if result.IsTrue() {
		return "TRUE         "
	}
	if result.IsFalse() {
		return "FALSE        "
	}
	return "INDETERMINATE"
}

func printConditions(tag string, report *engine.Report) {
	if report == nil {
		return
	}
	for _, route := range report.Routes() {
		var items []engine.Item
		switch tag {
		case "matched  ":
			if len(route.Unmatched()) == 0 && len(route.Absent()) == 0 {
				items = route.Matched()
			}
		case "unmatched":
			items = route.Unmatched()
		}
		for _, item := range items {
			if item.ConditionID() != "" {
				fmt.Printf("│      %s: %s (%s)\n", tag, item.ConditionName(), item.ConditionID())
			}
		}
	}
}
