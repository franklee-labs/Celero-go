package condition

import (
	"fmt"

	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

type AbsentCondition struct {
	core.BaseCondition
	field      string
	expression string
	program    cel.Program
	cfg        *core.ConditionConfig
}

func NewAbsentCondition(field string, cfg *core.ConditionConfig) *AbsentCondition {
	return &AbsentCondition{
		BaseCondition: core.NewBaseCondition(cfg),
		field:         field,
		expression:    fmt.Sprintf("!has( %s )", field),
		program:       nil,
		cfg:           cfg,
	}
}

func (c *AbsentCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	b, absent, err := Out(out, err)
	if absent {
		return true, false, nil
	}
	return b, false, err
}

func (c *AbsentCondition) Validate() core.Validation {
	if c.field == "" {
		return core.Invalid("field is empty")
	}
	return core.ValidationOK
}

func (c *AbsentCondition) Expression() string {
	return c.expression
}
func (c *AbsentCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return nil
}

func (c *AbsentCondition) Negate() (core.Condition, error) {
	n := NewExistsCondition(c.field, c.cfg)
	if v := n.Validate(); !v.Valid {
		return nil, &InvalidNodeError{"got invalid node after negate"}
	}
	err := n.Compile()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *AbsentCondition) Compile() error {
	return c.buildCELProgram(nil)
}

func (c *AbsentCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
