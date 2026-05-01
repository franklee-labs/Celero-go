package condition

import (
	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

// NegateCELCondition Logical inverse of CelCondition
type NegateCELCondition struct {
	core.BaseCondition
	origin     *CELCondition
	expression string
	program    cel.Program
}

func newNegateCELCondition(origin *CELCondition) *NegateCELCondition {
	return &NegateCELCondition{
		BaseCondition: core.NewBaseCondition(origin.cfg),
		origin:        origin,
		expression:    "!(" + origin.expression + ")",
	}
}

func (c *NegateCELCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	return Out(out, err)
}

func (c *NegateCELCondition) Validate() core.Validation {
	if c.origin.expression == "" {
		return core.Invalid("CEL expression is empty")
	}
	return core.ValidationOK
}

func (c *NegateCELCondition) Expression() string {
	return c.expression
}
func (c *NegateCELCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return nil
}

func (c *NegateCELCondition) Negate() (core.Condition, error) {
	return c.origin, nil
}

func (c *NegateCELCondition) Compile() error {
	return c.buildCELProgram(nil)
}

func (c *NegateCELCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
