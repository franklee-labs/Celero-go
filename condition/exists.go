package condition

import (
	"fmt"

	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

type ExistsCondition struct {
	core.BaseCondition
	field      string
	expression string
	program    cel.Program
	cfg        *core.ConditionConfig
}

func NewExistsCondition(field string, cfg *core.ConditionConfig) *ExistsCondition {
	return &ExistsCondition{
		BaseCondition: core.NewBaseCondition(cfg),
		field:         field,
		expression:    fmt.Sprintf("has( %s )", field),
		program:       nil,
		cfg:           cfg,
	}
}

func (c *ExistsCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	b, _, err := Out(out, err)
	return b, false, err
}

func (c *ExistsCondition) Validate() core.Validation {
	if c.field == "" {
		return core.Invalid("field is empty")
	}
	return core.ValidationOK
}

func (c *ExistsCondition) Expression() string {
	return c.expression
}
func (c *ExistsCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return nil
}

func (c *ExistsCondition) Negate() (core.Condition, error) {
	n := NewAbsentCondition(c.field, c.cfg)
	if v := n.Validate(); !v.Valid {
		return nil, &InvalidNodeError{"got invalid node after negate"}
	}
	err := n.Compile()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *ExistsCondition) Compile() error {
	return c.buildCELProgram(nil)
}

func (c *ExistsCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
