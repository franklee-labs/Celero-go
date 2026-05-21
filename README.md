# Celero-Go

<img src="./assets/celero_blue.svg" alt="Celero logo" width="250">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**celero-go** is a lightweight, easy-to-use Go rule engine for defining and evaluating complex business rules via API or JSON configuration.

## Features

- **Flexible rule definition**: build rules programmatically or deserialize from JSON
- **Rich condition types**: equality, comparison, regex, collection membership, list intersection/disjointness, field existence, CEL expressions, and more
- **Logical operators**: AND / OR / NOT, arbitrarily nestable
- **Three-valued logic**: `True` / `False` / `Indeterminate` result states (`AdvancedEngine`)
- **Cross-path condition result cache**: a shared condition across multiple paths is evaluated only once (opt-in)
- **Condition priority**: control the execution order of conditions within a path via `priority`
- **Event listeners**: rule-level and condition-level callbacks with ordering support
- **Evaluation reports**: per-path record of matched, unmatched, absent, and skipped conditions
- **CEL expression support**: integrates [Google CEL](https://github.com/google/cel-go) for advanced expression evaluation
- **Solutions API**: inspect the expanded path structure of a rule before evaluation

---

## Installation

```bash
go get github.com/franklee-labs/celero-go
```

---

## Core Design

### Rule Tree → Path Expansion

A Celero rule is described as a **logic tree**: leaf nodes are condition nodes, and internal nodes are relation nodes (AND / OR / NOT).

During `RuleBuilder.Build()`, the engine **expands the logic tree once** into a flat set of paths (a `PathGroup`). Each path is an ordered list of conditions; a path passes when every condition in it is satisfied.

Expansion rules:

| Expression | Expanded paths |
| --- | --- |
| `AND(A, B)` | Single path: `[A, B]` |
| `OR(A, B)` | Two paths: `[A]` and `[B]` |
| `AND(A, OR(B, C))` | Two paths: `[A, B]` and `[A, C]` |
| `NOT(AND(A, B))` | De Morgan expansion → `OR(NOT(A), NOT(B))` |

The engine iterates paths **in order**; the rule is true as soon as any path passes entirely. This design reduces the evaluation of complex logic to a simple sequential scan over flat lists.

### Three-Valued EvalResult

```go
// core package
var EvalResultTrue          = EvalResult{state: True}
var EvalResultFalse         = EvalResult{state: False}
var EvalResultIndeterminate = EvalResult{state: Indeterminate}

result.IsTrue()          // bool
result.IsFalse()         // bool
result.IsIndeterminate() // bool
```

- **True** — the condition/rule is definitively true
- **False** — the condition/rule is definitively false
- **Indeterminate** — a required parameter is absent from the context; the outcome cannot be determined

`Indeterminate` is only returned by `AdvancedEngine`. `DefaultEngine` treats missing parameters as `False` and always returns a plain `bool`.

### Two Engines

| | `DefaultEngine` | `AdvancedEngine` |
| --- | --- | --- |
| Return type | `(bool, error)` | `(core.EvalResult, error)` |
| Missing parameter handling | Treated as `False` | Returns `Indeterminate` |
| Event types | `ConditionEvent` / `RuleEvent` | `AdvancedConditionEvent` / `AdvancedRuleEvent` |

**Indeterminate propagation (`AdvancedEngine`)**:

Within a path, a `False` result short-circuits immediately. If all conditions are evaluated without any `False` but at least one is `Indeterminate`, the path result is `Indeterminate`.

At the rule level: if no path returns `True` but at least one path returned `Indeterminate`, the rule returns `Indeterminate`; otherwise it returns `False`.

```text
path1: [A=Indeterminate, B=True]  → Indeterminate
path2: [A=True, B=False]           → False (short-circuit)

rule → Indeterminate (no path is True, but one path is uncertain)
```

### Cross-Path Condition Result Cache

After expansion, the same condition node instance may appear in multiple paths. For example, `AND(A, OR(B, C))` expands to `[A, B]` and `[A, C]` — condition A appears in both paths.

Without caching, A is evaluated twice (especially costly for complex conditions such as regular expressions).

**With caching enabled**, the result of the first execution of A is stored in the evaluation context; subsequent paths read from the cache and skip re-evaluation.

Caching requires **both** switches to be enabled (independent, dual-gate design):

1. **Rule-level switch**: `builder.Cacheable(true)` — allows the rule to use caching at all
2. **Condition-level switch**: `"cacheable": true` on the condition node — allows this specific condition's result to be cached

```text
rule cacheable = false                                    →  no caching, regardless of condition setting
rule cacheable = true, condition cacheable = false        →  this condition is not cached
rule cacheable = true, condition cacheable = true         →  result is cached and reused across paths
```

Cache lifetime is **scoped to a single rule evaluation**; it never leaks across rules or requests.

> **Note**: `ConditionListener` is only fired when a condition is actually executed. A cache hit does not trigger the listener.

---

## Quick Start

### Building a Rule Programmatically

```go
import (
    "github.com/franklee-labs/celero-go/engine"
    "github.com/franklee-labs/celero-go/rules"
)

condNode := rules.ConditionNode{}
condNode.Sign = "EQ"
condNode.Props = map[string]interface{}{
    "id":        "cond-status",
    "name":      "Status Check",
    "field":     "status",
    "value":     "active",
    "valueType": "String",
    "priority":  0,
}

relNode := rules.CreateRelationNode("AND", condNode)

rule, err := engine.NewRuleBuilder().
    ID("rule-001").
    Name("Active User Check").
    Root(&relNode).
    Build()
if err != nil {
    panic(err)
}

eng := engine.NewDefaultEngine()
ctx := engine.NewRuleContext(map[string]interface{}{"status": "active"})

ok, err := eng.Evaluate(rule, ctx) // true
```

### Building a Rule from JSON

`FromJSON` accepts the rule id and the JSON of the logic tree root node. Condition properties are nested under a `"properties"` object. Fields such as `cacheable` and `ignoreAbsence` are **top-level** condition node fields, not inside `"properties"`. The `id`, `name`, and `priority` of a condition are placed **inside** `"properties"`.

```go
ruleJSON := `{
  "type": "relation",
  "sign": "AND",
  "children": [
    {
      "type": "condition",
      "sign": "GT",
      "cacheable": false,
      "ignoreAbsence": false,
      "properties": {
        "id": "age-cond",
        "name": "Age Check",
        "field": "age",
        "value": "18",
        "valueType": "Number",
        "priority": 0
      }
    }
  ]
}`

builder, err := engine.FromJSON("age-check", ruleJSON)
if err != nil {
    panic(err)
}
rule, err := builder.Name("Adult Verification").Build()
if err != nil {
    panic(err)
}

eng := engine.NewDefaultEngine()
ctx := engine.NewRuleContext(map[string]interface{}{"age": int64(25)})
ok, err := eng.Evaluate(rule, ctx) // true
```

### Composing Logic

```go
// AND(A, OR(B, C))
orNode := rules.CreateRelationNode("OR", condB, condC)
andNode := rules.CreateRelationNode("AND", condA, orNode)

rule, err := engine.NewRuleBuilder().ID("rule").Name("rule").Root(&andNode).Build()
```

This expands to two paths: `[A, B]` and `[A, C]`.

### Enabling the Cache (Multi-Path Scenario)

```go
// Mark condition A as cacheable at the node level
condA := rules.ConditionNode{}
condA.Sign = "REGEXP"
condA.Cacheable = true
condA.Props = map[string]interface{}{
    "id":       "cond-a",
    "name":     "Email Check",
    "field":    "email",
    "regexp":   `^[\w.+-]+@[\w-]+\.[a-z]{2,}$`,
    "priority": 0,
}

// Also enable the rule-level cache switch
orNode := rules.CreateRelationNode("OR", condB, condC)
andNode := rules.CreateRelationNode("AND", condA, orNode)

rule, err := engine.NewRuleBuilder().
    ID("rule").
    Cacheable(true). // rule-level switch
    Root(&andNode).
    Build()

// AND(A, OR(B, C)) → path1=[A,B], path2=[A,C]
// path1: evaluate A (result cached), evaluate B → false
// path2: read A's cached result (not re-executed), evaluate C → true
eng := engine.NewDefaultEngine()
ctx := engine.NewRuleContext(map[string]interface{}{"email": "alice@example.com", "level": "high"})
ok, err := eng.Evaluate(rule, ctx)
```

In JSON, set `cacheable` at the condition node level:

```json
{
  "type": "condition",
  "sign": "REGEXP",
  "cacheable": true,
  "ignoreAbsence": false,
  "properties": {
    "id": "cond-a",
    "name": "Email Check",
    "field": "email",
    "regexp": "^[\\w.+-]+@[\\w-]+\\.[a-z]{2,}$",
    "priority": 0
  }
}
```

---

## Using AdvancedEngine (Three-Valued Results)

```go
eng := engine.NewAdvancedEngine()

// All parameters present and matching → True
r1, _ := eng.Evaluate(rule, engine.NewRuleContext(map[string]interface{}{"status": "active"}))
r1.IsTrue() // true

// Parameters present but not matching → False
r2, _ := eng.Evaluate(rule, engine.NewRuleContext(map[string]interface{}{"status": "inactive"}))
r2.IsFalse() // true

// Required parameter missing → Indeterminate (cannot determine)
r3, _ := eng.Evaluate(rule, engine.NewRuleContext(map[string]interface{}{}))
r3.IsIndeterminate() // true
```

A typical use case for `Indeterminate`: in progressive rule-matching scenarios where data arrives incrementally, it lets you distinguish "definitively does not match" from "data is missing, cannot yet decide" — preventing false negatives. For example, when a user is filling out a form, the engine can match only the fields already provided; fields not yet entered return `Indeterminate` rather than `False`.

---

## Condition Reference

| Sign | Description | Properties |
| --- | --- | --- |
| `EQ` | Equal to | `field`, `value`, `valueType`: String / Number / Boolean / Expression |
| `NEQ` | Not equal to | `field`, `value`, `valueType`: String / Number / Boolean / Expression |
| `GT` | Greater than | `field`, `value`, `valueType`: Number / Expression |
| `GTE` | Greater than or equal | `field`, `value`, `valueType`: Number / Expression |
| `LT` | Less than | `field`, `value`, `valueType`: Number / Expression |
| `LTE` | Less than or equal | `field`, `value`, `valueType`: Number / Expression |
| `IN` | Value exists in collection | `field`, `value` (JSON array string), `valueType`: List |
| `NIN` | Value does not exist in collection | `field`, `value` (JSON array string), `valueType`: List |
| `REGEXP` | Regular expression match | `field`, `regexp` (regex pattern string) |
| `CEL` | Google CEL expression | `expression` (CEL expression string) |
| `INTERSECT` | Two lists share at least one common element | `field1`, `valueType1`, `field2`, `valueType2`: List / Expression |
| `DISJOINT` | Two lists share no common elements | `field1`, `valueType1`, `field2`, `valueType2`: List / Expression |
| `EXISTS` | Field is present in the evaluation context | `field` (supports dot notation, e.g. `params.name`) |
| `ABSENT` | Field is absent from the evaluation context | `field` (supports dot notation) |

### Notes on specific conditions

**`INTERSECT` / `DISJOINT`**: when `valueType` is `List`, the field value is a JSON array literal (e.g., `"[\"a\",\"b\"]"`); when `valueType` is `Expression`, the field is a context variable name holding a list.

**`EXISTS` / `ABSENT`**: always return a definite `True` or `False` regardless of context mode. `ABSENT` is the logical negation of `EXISTS`. `ignoreAbsence` does not apply to these two conditions.

**`REGEXP`**: the pattern is provided via the `regexp` property key (not `value`).

**Numeric types**: when the `value` string has no decimal part (e.g., `"18"`), the engine compiles it as `int64`; when it has a decimal part (e.g., `"18.5"`), it is compiled as `float64`. The corresponding parameter in `RuleContext` must match — pass `int64` for integer rules, `float64` for decimal rules.

### CEL Expression Example

```go
condNode := rules.ConditionNode{}
condNode.Sign = "CEL"
condNode.Props = map[string]interface{}{
    "id":         "cel-cond",
    "name":       "Age and Status",
    "expression": "age > 18 && status == 'active'",
    "priority":   0,
}
```

---

## Condition Priority

Within a path, a lower priority value means earlier execution (analogous to `ORDER BY priority ASC`):

```go
// core package constants
core.PriorityHighest = math.MinInt32  // executed first
core.PriorityDefault = 0
core.PriorityLowest  = math.MaxInt32  // executed last
```

Set `priority` inside the `properties` object of a condition node:

```json
{
  "type": "condition",
  "sign": "REGEXP",
  "cacheable": false,
  "ignoreAbsence": false,
  "properties": {
    "id":       "cond-1",
    "name":     "Email Format",
    "field":    "email",
    "regexp":   "^.+@.+$",
    "priority": 1
  }
}
```

---

## Ignoring Absence (`ignoreAbsence`)

By default, when a condition's required parameter is missing from the context:

- `DefaultEngine` → returns `False`
- `AdvancedEngine` → returns `Indeterminate`

Setting `ignoreAbsence = true` on a condition overrides this: a missing parameter **always** returns `False`, even in `AdvancedEngine`. This is useful for optional fields that should simply fail the condition rather than make the entire rule indeterminate.

`ignoreAbsence` is a **top-level field** on the condition node, not inside `properties`.

**Programmatic:**

```go
condNode := rules.ConditionNode{}
condNode.Sign = "EQ"
condNode.IgnoreAbsence = true // top-level node field, not in properties
condNode.Props = map[string]interface{}{
    "id":        "opt-cond",
    "name":      "Optional Tag Check",
    "field":     "optionalTag",
    "value":     "vip",
    "valueType": "String",
    "priority":  0,
}
```

**JSON:**

```json
{
  "type": "condition",
  "sign": "EQ",
  "cacheable": false,
  "ignoreAbsence": true,
  "properties": {
    "id":        "opt-cond",
    "name":      "Optional Tag Check",
    "field":     "optionalTag",
    "value":     "vip",
    "valueType": "String",
    "priority":  0
  }
}
```

`ignoreAbsence` is supported by: `EQ`, `NEQ`, `GT`, `GTE`, `LT`, `LTE`, `IN`, `NIN`, `REGEXP`, `CEL`, `INTERSECT`, `DISJOINT`.
It is **not** applicable to `EXISTS` / `ABSENT`, since those conditions are specifically about field presence.

---

## Event Listeners

Listeners are registered before evaluation and called in ascending `Order()` value (lower = earlier). Panics inside a listener do not affect rule evaluation.

### DefaultEngine

Implement the `ConditionListener` or `RuleListener` interface:

```go
type ConditionListener interface {
    OnResult(event *ConditionEvent)
    Order() int
}

type RuleListener interface {
    OnResult(event *RuleEvent)
    Order() int
}
```

```go
type myConditionListener struct{}

func (l *myConditionListener) OnResult(e *engine.ConditionEvent) {
    fmt.Printf("Condition %s matched: %v\n", e.ConditionName(), e.Matched())
}
func (l *myConditionListener) Order() int { return 0 }

eng := engine.NewDefaultEngine()
eng.AddConditionListener(&myConditionListener{})
eng.AddRuleListener(&myRuleListener{})
```

### AdvancedEngine

Use `AdvancedConditionListener` / `AdvancedRuleListener`; events carry a `core.EvalResult` (three-valued):

```go
type myAdvancedCondListener struct{}

func (l *myAdvancedCondListener) OnResult(e *engine.AdvancedConditionEvent) {
    result := e.Result() // core.EvalResult
    if result.IsIndeterminate() {
        fmt.Println("Missing field for:", e.ConditionName())
    }
}
func (l *myAdvancedCondListener) Order() int { return 0 }

eng := engine.NewAdvancedEngine()
eng.AddConditionListener(&myAdvancedCondListener{})
```

### RuleContext Attributes — sharing state between listeners

Listeners can read and write arbitrary key-value attributes on the `RuleContext` to share state within a single evaluation:

```go
type countListener struct{ order int }

func (l *countListener) OnResult(e *engine.ConditionEvent) {
    count := 0
    if v, ok := e.Context().GetAttribute("count"); ok {
        count = v.(int)
    }
    e.Context().SetAttribute("count", count+1)
}
func (l *countListener) Order() int { return l.order }

eng := engine.NewDefaultEngine()
eng.AddConditionListener(&countListener{order: 1})

ctx := engine.NewRuleContext(params)
eng.Evaluate(rule, ctx)

total, _ := ctx.GetAttribute("count")
fmt.Println("Total:", total)
```

---

## Evaluation Reports

When reports are enabled, each rule evaluation records the status of every condition on every path:

```go
ctx := engine.NewRuleContext(params)
ctx.EnableReports()

eng.Evaluate(rule, ctx)

reports := ctx.Reports() // map[*engine.CeleroRule]*engine.Report
for rule, report := range reports {
    for _, route := range report.Routes() {
        route.Matched()   // conditions that evaluated to true
        route.Unmatched() // conditions that evaluated to false (caused path to fail)
        route.Absent()    // conditions with Indeterminate result (AdvancedEngine only)
        route.Skipped()   // conditions not evaluated due to short-circuit
    }
}
```

Each `Route` corresponds to one path evaluation attempt and contains `Item` values (conditionID + conditionName).

---

## Batch Evaluation

```go
rules := []*engine.CeleroRule{rule1, rule2, rule3}
ctx := engine.NewRuleContext(params)

// DefaultEngine — fires RuleListener after each rule
eng.Evalutes(rules, ctx)

// AdvancedEngine — same, but RuleListener receives core.EvalResult
advEng.Evalutes(rules, ctx)
```

> **Note**: the single-rule overload `eng.Evaluate(rule, ctx)` does **not** fire `RuleListener`. Use the batch overload when you need rule-level callbacks.

---

## Solutions API

`CeleroRule.Solutions()` exposes the expanded path structure for inspection before or after evaluation:

```go
solutions := rule.Solutions()
fmt.Println("Number of paths:", solutions.Count())

sol, ok := solutions.SolutionAt(0)
if ok {
    fmt.Println("Conditions in path 0:", sol.Count())
    cond, _ := sol.ConditionAt(0)
    fmt.Println("First condition ID:", cond.ID())
}
```

---

## Custom Condition Types

Register a custom creator globally (available to all rules) or scoped to a specific rule:

```go
// Implement ConditionCreator
type StartsWithCreator struct{}

func (c *StartsWithCreator) Create(node rules.ConditionNode) (core.Condition, error) {
    field, _ := rules.StringValue(node.Properties(), "field")
    prefix, _ := rules.StringValue(node.Properties(), "value")
    return NewStartsWithCondition(field, prefix), nil
}

// Register globally — sign must not conflict with built-in signs
rules.RegisterConditionCreator("STARTS_WITH", &StartsWithCreator{})

// Or scope it to a specific rule
rules.RegisterRuleConditionCreator("my-rule-id", "STARTS_WITH", &StartsWithCreator{})
```

Then use the sign in a rule definition like any built-in sign.

---

## Architecture Overview

```text
RuleBuilder
  ├── FromJSON(id, jsonStr) / NewRuleBuilder()
  └── Build()
       ├── RelationNode.Transform()  → logic tree (AND / OR / NOT + Condition nodes)
       ├── core.ValidateAll()        → structural validation
       └── Relation.Resolve()        → expand into PathGroup (list of Paths)
                                          each Path sorted by condition priority

Evaluation:
  Engine.Evaluate(rule, ruleContext)
    └── iterate each Path in PathGroup
         └── execute each Condition in order
              ├── check cache (if enabled)
              ├── Condition.Evaluate(context)
              │    ├── evaluate() succeeds → True / False
              │    └── absent
              │         ├── ignoreAbsence=true          → False
              │         ├── DefaultEngine               → False
              │         └── AdvancedEngine              → Indeterminate
              └── write to cache (if cacheable)
```

### Package Structure

```text
github.com/franklee-labs/celero-go/
├── core/        EvalResult, Condition, Relation, Node, Path, PathGroup,
│                ConditionConfig, Priority constants, ValueType, Validation
├── engine/      DefaultEngine, AdvancedEngine, CeleroRule, RuleBuilder,
│                RuleContext, Report, Route, Solutions, listeners
├── rules/       RelationNode, ConditionNode, ConditionCreator,
│                RegisterConditionCreator, RegisterRuleConditionCreator
├── condition/   EqualCondition, CompareCondition, CelCondition, RegexpCondition,
│                InCondition, IntersectCondition, DisjointCondition,
│                ExistsCondition, AbsentCondition, ...
└── logic/       AND, OR, NOT relation implementations
```

---

## Testing

```bash
go test ./...
```

Test coverage includes:

- All condition types (equality, comparison, regex, CEL, list intersection/disjointness, field existence)
- AND / OR / NOT logic with arbitrary nesting
- Three-valued (Indeterminate) propagation and edge cases
- All cache switch combinations (rule-level × condition-level × cache hit/miss)
- Evaluation reports (matched / unmatched / absent / skipped)
- JSON rule parsing and programmatic rule building
- Solutions API bounds checking

---

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a branch (`git checkout -b feature/xxx`)
3. Commit your changes (`git commit -m 'Add xxx'`)
4. Push (`git push origin feature/xxx`)
5. Open a Pull Request

---

## License

[MIT](LICENSE)

---

Made with ❤️ by [franklee-labs](https://github.com/franklee-labs)
