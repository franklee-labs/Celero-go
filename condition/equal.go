package condition

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

type EqualCondition struct {
	core.BaseCondition
	field         string
	value         string
	valueType     core.ValueType
	sign          string
	expression    string
	builtinParams map[string]interface{}
	program       cel.Program
	cfg           *core.ConditionConfig
}

func NewEqualCondition(field, value string, valueType core.ValueType, cfg *core.ConditionConfig) *EqualCondition {
	c := &EqualCondition{
		BaseCondition: core.NewBaseCondition(cfg),
		field:         field,
		value:         value,
		valueType:     valueType,
		sign:          " == ",
		expression:    "",
		builtinParams: nil,
		program:       nil,
		cfg:           cfg,
	}
	if c.Name() == "" {
		c.SetName(c.generateName())
	}
	return c
}

func (c *EqualCondition) Validate() core.Validation {
	if c.valueType == core.ValueTypeNumber {
		_, err := strconv.ParseFloat(c.value, 64)
		if err != nil {
			return core.Invalid(fmt.Sprintf("value is not a number err:%s", err.Error()))
		}
	}
	return core.ValidationOK
}

func (c *EqualCondition) Expression() string {
	return c.expression
}

func (c *EqualCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return ctx.BuildEvalParams(c.builtinParams)
}

func (c *EqualCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	return Out(out, err)
}

func (c *EqualCondition) Negate() (core.Condition, error) {
	n := NewNotEqualCondition(c.field, c.value, c.valueType, c.cfg)
	if v := n.Validate(); !v.Valid {
		return nil, &InvalidNodeError{"got invalid node after negate"}
	}
	err := n.Compile()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *EqualCondition) Compile() error {
	switch c.valueType {
	case core.ValueTypeExpression:
		c.expression = c.field + c.sign + c.value
		return c.buildCELProgram(nil)
	case core.ValueTypeString:
		innerKey := builtin_inner_key_prefix + "STR_001"
		valKey := builtin_var + "." + innerKey
		c.expression = c.field + c.sign + valKey
		c.builtinParams = mapOf(builtin_var, mapOf(innerKey, c.value))
		return c.buildCELProgram(mapOfCelType(builtin_var, cel.MapType(cel.StringType, cel.StringType)))
	case core.ValueTypeBool:
		v, err := strconv.ParseBool(c.value)
		if err != nil {
			return err
		}
		if v {
			c.expression = c.field + c.sign + "true"
		} else {
			c.expression = c.field + c.sign + "false"
		}
		return c.buildCELProgram(nil)
	case core.ValueTypeNumber:
		innerKey := builtin_inner_key_prefix + "NUM_001"
		valKey := builtin_var + "." + innerKey
		c.expression = c.field + c.sign + valKey
		// Try exact integer parse first ("42"), then float whole-number ("42.0").
		if i, err := strconv.ParseInt(c.value, 10, 64); err == nil {
			c.builtinParams = mapOf(builtin_var, mapOf(innerKey, i))
			return c.buildCELProgram(mapOfCelType(builtin_var, cel.MapType(cel.StringType, cel.IntType)))
		}
		f, err := strconv.ParseFloat(c.value, 64)
		if err != nil {
			return fmt.Errorf("invalid number value %q: %w", c.value, err)
		}
		if f == math.Trunc(f) && !math.IsInf(f, 0) {
			c.builtinParams = mapOf(builtin_var, mapOf(innerKey, int64(f)))
			return c.buildCELProgram(mapOfCelType(builtin_var, cel.MapType(cel.StringType, cel.IntType)))
		}
		c.builtinParams = mapOf(builtin_var, mapOf(innerKey, f))
		return c.buildCELProgram(mapOfCelType(builtin_var, cel.MapType(cel.StringType, cel.DoubleType)))
	default:
		return errors.New("Unsupported ValueType in EqualCondition.")
	}
}

func (c *EqualCondition) generateName() string {
	switch c.valueType {
	case core.ValueTypeString:
		return fmt.Sprintf("%s %s str(%s)", c.field, c.sign, c.value)
	case core.ValueTypeNumber:
		return fmt.Sprintf("%s %s num(%s)", c.field, c.sign, c.value)
	case core.ValueTypeExpression:
		return fmt.Sprintf("%s %s %s", c.field, c.sign, c.value)
	case core.ValueTypeBool:
		return fmt.Sprintf("%s %s bool(%s)", c.field, c.sign, c.value)
	default:
		return fmt.Sprintf("%s %s %s", c.field, c.sign, c.value)
	}
}

func (c *EqualCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
