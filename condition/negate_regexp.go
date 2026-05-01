package condition

import (
	"regexp"
	"strings"

	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

type NegateRegexpCondition struct {
	core.BaseCondition
	origin     *RegexpCondition
	regexp     string
	expression string
	program    cel.Program
	cfg        *core.ConditionConfig
}

func newNegateRegexpCondition(origin *RegexpCondition) *NegateRegexpCondition {
	return &NegateRegexpCondition{
		BaseCondition: core.NewBaseCondition(origin.cfg),
		origin:        origin,
		regexp:        origin.Regexp(),
		expression:    "!(" + origin.expression + ")",
		program:       nil,
		cfg:           origin.cfg,
	}
}

func (c *NegateRegexpCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	return Out(out, err)
}

func (c *NegateRegexpCondition) Validate() core.Validation {
	if strings.TrimSpace(c.regexp) == "" {
		return core.Invalid("regexp is empty")
	}
	_, err := regexp.Compile(c.regexp)
	if err != nil {
		return core.Invalid("regexp is invalid. " + err.Error())
	}
	return core.ValidationOK
}

func (c *NegateRegexpCondition) Expression() string {
	return c.expression
}
func (c *NegateRegexpCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return nil
}

func (c *NegateRegexpCondition) Negate() (core.Condition, error) {
	return c.origin, nil
}

func (c *NegateRegexpCondition) Compile() error {
	return c.buildCELProgram(nil)
}

func (c *NegateRegexpCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
