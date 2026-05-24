// Demonstrates condition result caching in a rule with the shape A AND (B OR C).
//
// The engine expands A AND (B OR C) into two paths:
//
//	path 1: A AND B
//	path 2: A AND C
//
// When A is cacheable and the rule has caching enabled:
//   - path 1 evaluates A (result cached), then B (false) → path fails
//   - path 2 reads A from cache (not re-executed), evaluates C (true) → path succeeds
//
// Two runs compare behavior:
//
//	runWithCache()    — A executed once,  B once, C once (cache hit on path 2)
//	runWithoutCache() — A executed twice, B once, C once (no cache, A re-evaluated)
package main

import (
	"fmt"

	"github.com/franklee-labs/celero-go/engine"
)

// A AND (B OR C)
// A: email matches a pattern  — REGEXP, cacheable=true
// B: role == "admin"          — EQ,     cacheable=true
// C: level == "high"          — EQ,     cacheable=true
const ruleJSON = `{
  "type": "relation",
  "sign": "AND",
  "children": [
    {
      "type": "condition",
      "sign": "REGEXP",
      "cacheable": true,
      "properties": {
        "id": "cond-a",
        "name": "A: email format",
        "priority": 0,
        "field": "email",
        "regexp": "^[\\w.+-]+@[\\w-]+\\.[a-z]{2,}$"
      }
    },
    {
      "type": "relation",
      "sign": "OR",
      "children": [
        {
          "type": "condition",
          "sign": "EQ",
          "cacheable": true,
          "properties": {
            "id": "cond-b",
            "name": "B: role check",
            "priority": 0,
            "field": "role",
            "value": "admin",
            "valueType": "String"
          }
        },
        {
          "type": "condition",
          "sign": "EQ",
          "cacheable": true,
          "properties": {
            "id": "cond-c",
            "name": "C: level check",
            "priority": 0,
            "field": "level",
            "value": "high",
            "valueType": "String"
          }
        }
      ]
    }
  ]
}`

// A=true (email valid), B=false (not admin), C=true (level=high)
// Expected evaluation order:
//
//	path 1 (A&B): A executes → true (cached),  B executes → false (path fails)
//	path 2 (A&C): A from cache → true (skipped), C executes → true (path succeeds)
var testUser = map[string]interface{}{
	"email": "alice@example.com",
	"role":  "user",
	"level": "high",
}

type executionTracker struct {
	counts map[string]int
	order  []string
}

func newExecutionTracker() *executionTracker {
	return &executionTracker{counts: make(map[string]int)}
}

func (t *executionTracker) OnResult(event *engine.ConditionEvent) {
	key := event.ConditionID() + " (" + event.ConditionName() + ")"
	if _, exists := t.counts[key]; !exists {
		t.order = append(t.order, key)
	}
	t.counts[key]++
}

func (t *executionTracker) Order() int { return 0 }

func (t *executionTracker) printSummary() {
	fmt.Println("  Condition execution counts:")
	for _, key := range t.order {
		fmt.Printf("    %-42s executed %d time(s)\n", key, t.counts[key])
	}
}

func runWithCache() {
	fmt.Println("── With caching (rule.cacheable=true) ────────────────")

	rb, err := engine.FromJSON("rule-cached", ruleJSON)
	if err != nil {
		panic(err)
	}
	rule, err := rb.Name("cached rule").Cacheable(true).Build()
	if err != nil {
		panic(err)
	}

	eng := engine.NewDefaultEngine()
	tracker := newExecutionTracker()
	eng.AddConditionListener(tracker)

	result, err := eng.Evaluate(rule, engine.NewRuleContext(testUser))
	if err != nil {
		panic(err)
	}
	fmt.Printf("  Rule matched: %v\n", result)
	tracker.printSummary()
	fmt.Println("  ✓ A evaluated once: path 2 read the cached result instead of re-executing.")
	fmt.Println()
}

func runWithoutCache() {
	fmt.Println("── Without caching (rule.cacheable=false) ────────────")

	rb, err := engine.FromJSON("rule-uncached", ruleJSON)
	if err != nil {
		panic(err)
	}
	rule, err := rb.Name("uncached rule").Cacheable(false).Build()
	if err != nil {
		panic(err)
	}

	eng := engine.NewDefaultEngine()
	tracker := newExecutionTracker()
	eng.AddConditionListener(tracker)

	result, err := eng.Evaluate(rule, engine.NewRuleContext(testUser))
	if err != nil {
		panic(err)
	}
	fmt.Printf("  Rule matched: %v\n", result)
	tracker.printSummary()
	fmt.Println("  ✓ A evaluated twice: no cache, so path 2 re-executed A from scratch.")
	fmt.Println()
}

func main() {
	fmt.Println("Rule shape:  A AND (B OR C)")
	fmt.Println("Paths:       [A AND B]  [A AND C]")
	fmt.Println("Input:       A=true (email valid)  B=false (not admin)  C=true (level=high)")
	fmt.Println()

	runWithCache()
	runWithoutCache()
}
