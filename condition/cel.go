package condition

import (
	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

// CELCondition evaluates a CEL expression against the rule context params.
// The expression must return a bool, e.g. "params.age >= 18".
type CELCondition struct {
	core.BaseCondition
	expression string
	program    cel.Program
	cfg        *core.ConditionConfig
}

func NewCELCondition(expression string, cfg *core.ConditionConfig) *CELCondition {
	return &CELCondition{
		BaseCondition: core.NewBaseCondition(cfg),
		expression:    expression,
		program:       nil,
		cfg:           cfg,
	}
}

// Execute evaluates the CEL expression with params from ctx.
// Returns Indeterminate on error or non-bool result.
func (c *CELCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	return Out(out, err)
}

func (c *CELCondition) Validate() core.Validation {
	if c.expression == "" {
		return core.Invalid("CEL expression is empty")
	}
	return core.ValidationOK
}

func (c *CELCondition) Expression() string {
	return c.expression
}
func (c *CELCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return nil
}

func (c *CELCondition) Negate() (core.Condition, error) {
	n := newNegateCELCondition(c)
	if v := n.Validate(); !v.Valid {
		return nil, &InvalidNodeError{"got invalid node after negate"}
	}
	err := n.Compile()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *CELCondition) Compile() error {
	return c.buildCELProgram(nil)
}

func (c *CELCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
