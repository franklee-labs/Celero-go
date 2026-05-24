// Demonstrates how ignoreAbsence interacts with INDETERMINATE inside the same rule.
//
// Rules (coupon-rules-ignore-absence.json) mix per-condition ignoreAbsence settings:
//
//	high-value-user:
//	  c-spend  ignoreAbsence=true  — missing totalSpend  → FALSE  (collapse to false)
//	  c-level  ignoreAbsence=false — missing memberLevel → INDETERMINATE
//
//	vip-access:
//	  c-age      ignoreAbsence=true  — missing age      → FALSE  (collapse to false)
//	  c-verified ignoreAbsence=false — missing verified → INDETERMINATE
//
// Decision path when a field is absent:
//
//	ignoreAbsence=true  → FALSE          (field treated as not matching, result is certain)
//	ignoreAbsence=false → INDETERMINATE  (only when AdvancedEngine / enableMissing=true)
//
// Test cases:
//
//	Frank1 — missing totalSpend  (ignoreAbsence=true)  → high-value-user = FALSE
//	Frank2 — missing memberLevel (ignoreAbsence=false) → high-value-user = INDETERMINATE
//	Grace  — missing age + verified
//	             age      ignoreAbsence=true  → c-age=FALSE
//	             verified ignoreAbsence=false → c-verified=INDETERMINATE
//	             OR(FALSE, INDETERMINATE)     → INDETERMINATE
//	             vip-access = INDETERMINATE
package main

import (
	"fmt"

	"github.com/franklee-labs/celero-go/core"
	"github.com/franklee-labs/celero-go/engine"
	"github.com/franklee-labs/celero-go/examples/shared"
)

func main() {
	rules, err := shared.LoadCouponRulesIgnoreAbsence()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded rules: %d\n", len(rules))
	fmt.Println("  high-value-user: c-spend(ignoreAbsence=true)  c-level(ignoreAbsence=false)")
	fmt.Println("  vip-access:      c-age(ignoreAbsence=true)    c-verified(ignoreAbsence=false)")
	fmt.Println()

	eng := engine.NewAdvancedEngine()

	users := []map[string]interface{}{
		// all fields present — same result as advanced_engine example
		{"name": "Alice", "totalSpend": int64(800), "memberLevel": "gold", "registeredDays": int64(365), "orderCount": int64(20), "status": "active", "age": int64(25), "verified": true, "banned": false},
		// missing totalSpend: c-spend ignoreAbsence=true → FALSE (short-circuits AND)
		// expected: high-value-user = FALSE
		{"name": "Frank1 (missing totalSpend)", "memberLevel": "gold", "registeredDays": int64(180), "orderCount": int64(10), "status": "active", "age": int64(28), "verified": true, "banned": false},
		// missing memberLevel: c-level ignoreAbsence=false → INDETERMINATE
		// expected: high-value-user = INDETERMINATE
		{"name": "Frank2 (missing memberLevel)", "totalSpend": int64(800), "registeredDays": int64(180), "orderCount": int64(10), "status": "active", "age": int64(28), "verified": true, "banned": false},
		// missing age + verified:
		//   c-age ignoreAbsence=true → FALSE, c-verified ignoreAbsence=false → INDETERMINATE
		//   OR(FALSE, INDETERMINATE) → INDETERMINATE
		// expected: vip-access = INDETERMINATE
		{"name": "Grace (missing age + verified)", "totalSpend": int64(100), "memberLevel": "normal", "registeredDays": int64(60), "orderCount": int64(3), "status": "active", "banned": false},
	}

	for _, user := range users {
		fmt.Printf("┌── User: %s\n", user["name"])

		ctx := engine.NewRuleContext(user)
		for _, r := range rules {
			result, err := eng.Evaluate(r, ctx)
			if err != nil {
				panic(err)
			}
			fmt.Printf("│  %s [%s]\n", label(result), r.ID())
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
