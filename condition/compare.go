package condition

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

type CompareCondition struct {
	core.BaseCondition
	field         string
	value         string
	valueType     core.ValueType
	signStr       string
	sign          string
	expression    string
	builtinParams map[string]interface{}
	program       cel.Program
	cfg           *core.ConditionConfig
}

func negateSignStr(signStr string) string {
	signStr = strings.TrimSpace(signStr)
	switch signStr {
	case "GT":
		return "LTE"
	case "GTE":
		return "LT"
	case "LT":
		return "GTE"
	case "LTE":
		return "GT"
	default:
		return signStr
	}
}

func toSign(signStr string) string {
	signStr = strings.TrimSpace(signStr)
	switch signStr {
	case "GT":
		return " > "
	case "GTE":
		return " >= "
	case "LT":
		return " < "
	case "LTE":
		return " <= "
	default:
		return " " + signStr + " "
	}
}

func checkSign(sign string) bool {
	sign = strings.TrimSpace(sign)
	return sign == ">" || sign == ">=" || sign == "<" || sign == "<="
}

func NewCompareCondition(field, value, signStr string, valueType core.ValueType, cfg *core.ConditionConfig) *CompareCondition {
	c := &CompareCondition{
		BaseCondition: core.NewBaseCondition(cfg),
		field:         field,
		value:         value,
		valueType:     valueType,
		signStr:       signStr,
		sign:          toSign(signStr),
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

func (c *CompareCondition) Validate() core.Validation {
	b := checkSign(c.sign)
	if !b {
		return core.Invalid(fmt.Sprintf("unsupported sign: %s", c.sign))
	}
	switch c.valueType {
	case core.ValueTypeNumber:
		_, err := strconv.ParseFloat(c.value, 64)
		if err != nil {
			return core.Invalid(fmt.Sprintf("value is not a number err:%s", err.Error()))
		}
		return core.ValidationOK
	case core.ValueTypeExpression:
		return core.ValidationOK
	default:
		return core.Invalid(fmt.Sprintf("%s unsupported value type", c.Name()))
	}
}

func (c *CompareCondition) Expression() string {
	return c.expression
}

func (c *CompareCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return ctx.BuildEvalParams(c.builtinParams)
}

func (c *CompareCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	return Out(out, err)
}

func (c *CompareCondition) Negate() (core.Condition, error) {
	n := NewCompareCondition(c.field, c.value, negateSignStr(c.signStr), c.valueType, c.cfg)
	if v := n.Validate(); !v.Valid {
		return nil, &InvalidNodeError{"got invalid node after negate. " + v.Message}
	}
	err := n.Compile()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *CompareCondition) Compile() error {
	switch c.valueType {
	case core.ValueTypeExpression:
		c.expression = c.field + c.sign + c.value
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
		return errors.New("Unsupported ValueType in CompareCondition.")
	}
}

func (c *CompareCondition) generateName() string {
	if c.valueType == core.ValueTypeNumber {
		return fmt.Sprintf("%s %s num(%s)", c.field, c.sign, c.value)
	}
	return fmt.Sprintf("%s %s %s", c.field, c.sign, c.value)
}

func (c *CompareCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
