package condition

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

type RegexpCondition struct {
	core.BaseCondition
	field      string
	regexp     string
	expression string
	program    cel.Program
	cfg        *core.ConditionConfig
}

func NewRegexpCondition(field, regex string, cfg *core.ConditionConfig) *RegexpCondition {
	return &RegexpCondition{
		BaseCondition: core.NewBaseCondition(cfg),
		field:         field,
		regexp:        regex,
		expression:    "",
		program:       nil,
		cfg:           cfg,
	}
}

func (c *RegexpCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	return Out(out, err)
}

func (c *RegexpCondition) Validate() core.Validation {
	if strings.TrimSpace(c.regexp) == "" {
		return core.Invalid("regexp is empty")
	}
	_, err := regexp.Compile(c.regexp)
	if err != nil {
		return core.Invalid("regexp is invalid. " + err.Error())
	}
	return core.ValidationOK
}

func (c *RegexpCondition) Expression() string {
	return c.expression
}
func (c *RegexpCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return nil
}

func (c *RegexpCondition) Negate() (core.Condition, error) {
	n := newNegateRegexpCondition(c)
	if v := n.Validate(); !v.Valid {
		return nil, &InvalidNodeError{"got invalid node after negate"}
	}
	err := n.Compile()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *RegexpCondition) Compile() error {
	c.expression = fmt.Sprintf("%s.matches(r'%s')", c.field, c.regexp)
	return c.buildCELProgram(nil)
}

func (c *RegexpCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}

func (c *RegexpCondition) Regexp() string {
	return c.regexp
}
