package condition

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

type InCondition struct {
	core.BaseCondition
	field         string
	value         string
	valueType     core.ValueType
	slice         []interface{}
	expression    string
	sign          string
	initErr       error
	builtinParams map[string]interface{}
	program       cel.Program
	cfg           *core.ConditionConfig
}

func NewInCondition(field, value string, valueType core.ValueType, cfg *core.ConditionConfig) *InCondition {
	c := &InCondition{
		BaseCondition: core.NewBaseCondition(cfg),
		field:         field,
		value:         value,
		valueType:     valueType,
		slice:         nil,
		expression:    "",
		sign:          " in ",
		initErr:       nil,
		builtinParams: nil,
		program:       nil,
		cfg:           cfg,
	}
	if c.Name() == "" {
		c.SetName(c.generateName())
	}
	c.initErr = c.parseSlice()
	return c
}

func (c *InCondition) parseSlice() error {
	if strings.TrimSpace(c.value) == "" {
		return fmt.Errorf("value must be a JSON array. got: %s", c.value)
	}
	if c.valueType == core.ValueTypeList {
		var v []interface{}
		if err := json.Unmarshal([]byte(c.value), &v); err != nil {
			return err
		}
		c.slice = v
		return nil
	}
	return nil
}

func (c *InCondition) Validate() core.Validation {
	if c.initErr != nil {
		return core.Invalid("invalid value. " + c.initErr.Error())
	}
	switch c.valueType {
	case core.ValueTypeList:
		if len(c.slice) == 0 {
			return core.Invalid("value is nil or empty")
		}
		return core.ValidationOK
	case core.ValueTypeExpression:
		return core.ValidationOK
	default:
		return core.Invalid(fmt.Sprintf("%s unsupported value type, supported [List, Expression]", c.Name()))
	}
}

func (c *InCondition) Expression() string {
	return c.expression
}

func (c *InCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return ctx.BuildEvalParams(c.builtinParams)
}

func (c *InCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	return Out(out, err)
}

func (c *InCondition) Negate() (core.Condition, error) {
	n := NewNotInCondition(c.field, c.value, c.valueType, c.cfg)
	if v := n.Validate(); !v.Valid {
		return nil, &InvalidNodeError{"got invalid node after negate. " + v.Message}
	}
	err := n.Compile()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *InCondition) Compile() error {
	switch c.valueType {
	case core.ValueTypeExpression:
		c.expression = c.field + c.sign + c.value
		return c.buildCELProgram(nil)
	case core.ValueTypeList:
		innerKey := builtin_inner_key_prefix + "LIST_001"
		valKey := builtin_var + "." + innerKey
		c.expression = c.field + c.sign + valKey
		c.builtinParams = mapOf(builtin_var, mapOf(innerKey, c.slice))
		return c.buildCELProgram(mapOfCelType(builtin_var, cel.MapType(cel.StringType, cel.ListType(cel.DynType))))
	default:
		return errors.New("Unsupported ValueType in InCondition.")
	}
}

func (c *InCondition) generateName() string {
	return fmt.Sprintf("!( %s %s %s )", c.field, c.sign, c.value)
}

func (c *InCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
