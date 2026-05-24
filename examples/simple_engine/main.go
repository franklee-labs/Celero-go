// Demonstrates loading rules from JSON, creating a DefaultEngine, and evaluating results.
// Rule file: examples/shared/data/coupon-rules.json
// Scenario: determine which coupon each user qualifies for.
package main

import (
	"fmt"

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
	fmt.Println()
	runMultiRules(rules)
}

func runSingleRule(rules []*engine.CeleroRule) {
	eng := engine.NewDefaultEngine()

	fmt.Println("=== Single rule evaluation ===")
	var vipRule *engine.CeleroRule
	for _, r := range rules {
		if r.ID() == "vip-access" {
			vipRule = r
			break
		}
	}

	testUser := map[string]interface{}{
		"status": "active", "age": int64(17), "verified": true, "banned": false,
	}
	result, err := eng.Evaluate(vipRule, engine.NewRuleContext(testUser))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Rule [%s] matched: %v\n", vipRule.Name(), result)
}

func runMultiRules(rules []*engine.CeleroRule) {
	eng := engine.NewDefaultEngine()

	users := []map[string]interface{}{
		// expected: high-value-user
		{"name": "Alice", "totalSpend": int64(800), "memberLevel": "gold", "registeredDays": int64(365), "orderCount": int64(20), "status": "active", "age": int64(25), "verified": true, "banned": false},
		// expected: new-user-welcome
		{"name": "Bob", "totalSpend": int64(0), "memberLevel": "normal", "registeredDays": int64(7), "orderCount": int64(0), "status": "active", "age": int64(20), "verified": false, "banned": false},
		// expected: vip-access (adult, not banned)
		{"name": "Carol", "totalSpend": int64(200), "memberLevel": "silver", "registeredDays": int64(90), "orderCount": int64(5), "status": "active", "age": int64(22), "verified": false, "banned": false},
		// expected: high-value-user only (vip-access rejected — account inactive)
		{"name": "Dave", "totalSpend": int64(1000), "memberLevel": "platinum", "registeredDays": int64(200), "orderCount": int64(50), "status": "inactive", "age": int64(30), "verified": true, "banned": false},
		// expected: vip-access (minor but verified)
		{"name": "Eve", "totalSpend": int64(50), "memberLevel": "normal", "registeredDays": int64(15), "orderCount": int64(0), "status": "active", "age": int64(16), "verified": true, "banned": false},
	}

	for _, user := range users {
		fmt.Printf("┌── User: %s\n", user["name"])

		ctx := engine.NewRuleContext(user)
		ctx.EnableReports()
		eng.Evalutes(rules, ctx)

		for _, r := range rules {
			report := ctx.Reports()[r]
			if isMatched(report) {
				fmt.Printf("│  ✓ matched: [%s] %s\n", r.ID(), r.Name())
				printMatchedConditions(report)
			} else {
				fmt.Printf("│  ✗ [%s] unmatched conditions:\n", r.ID())
				printUnmatchedConditions(report)
			}
		}
		fmt.Println("└───────────────────────────────────────")
		fmt.Println()
	}
}

func isMatched(report *engine.Report) bool {
	if report == nil {
		return false
	}
	for _, route := range report.Routes() {
		if len(route.Unmatched()) == 0 && len(route.Absent()) == 0 {
			return true
		}
	}
	return false
}

func printMatchedConditions(report *engine.Report) {
	if report == nil {
		return
	}
	for _, route := range report.Routes() {
		if len(route.Unmatched()) == 0 && len(route.Absent()) == 0 {
			for _, item := range route.Matched() {
				if item.ConditionID() != "" {
					fmt.Printf("│      condition met: %s (%s)\n", item.ConditionName(), item.ConditionID())
				}
			}
		}
	}
}

func printUnmatchedConditions(report *engine.Report) {
	if report == nil {
		return
	}
	for _, route := range report.Routes() {
		for _, item := range route.Unmatched() {
			fmt.Printf("│      not satisfied: %s (%s)\n", item.ConditionName(), item.ConditionID())
		}
	}
}
