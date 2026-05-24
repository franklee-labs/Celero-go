package condition

import (
	"encoding/json"
	"fmt"

	"github.com/franklee-labs/celero-go/core"
	"github.com/google/cel-go/cel"
)

type DisjointCondition struct {
	core.BaseCondition
	field1        string
	valueTypeStr1 string
	valueType1    core.ValueType
	field2        string
	valueTypeStr2 string
	valueType2    core.ValueType
	initErr       error
	expression    string
	builtinParams map[string]interface{}
	program       cel.Program
	cfg           *core.ConditionConfig
}

func NewDisjointCondition(field1, valueType1, field2, valueType2 string, cfg *core.ConditionConfig) *DisjointCondition {
	c := &DisjointCondition{
		BaseCondition: core.NewBaseCondition(cfg),
		field1:        field1,
		valueTypeStr1: valueType1,
		field2:        field2,
		valueTypeStr2: valueType2,
		initErr:       nil,
		expression:    "",
		builtinParams: nil,
		program:       nil,
		cfg:           cfg,
	}
	if c.Name() == "" {
		c.SetName(c.generateName())
	}
	if err := c.setValueType(valueType1, valueType2); err != nil {
		c.initErr = err
	}
	return c
}

func (c *DisjointCondition) setValueType(valueType1, valueType2 string) error {
	vt1 := core.ValueTypeFromString(valueType1)
	if vt1 == core.ValueTypeInvalid {
		return fmt.Errorf("invalid ValueType.")
	}
	c.valueType1 = vt1
	vt2 := core.ValueTypeFromString(valueType2)
	if vt2 == core.ValueTypeInvalid {
		return fmt.Errorf("invalid ValueType.")
	}
	c.valueType2 = vt2
	return nil
}

func (c *DisjointCondition) Validate() core.Validation {
	if c.initErr != nil {
		return core.Invalid(c.initErr.Error())
	}
	if c.valueType1 != core.ValueTypeList && c.valueType1 != core.ValueTypeExpression {
		return core.Invalid("invalid value type. support [List, Expression]")
	}
	if c.valueType2 != core.ValueTypeList && c.valueType2 != core.ValueTypeExpression {
		return core.Invalid("invalid value type. support [List, Expression]")
	}
	return core.ValidationOK
}

func (c *DisjointCondition) Expression() string {
	return c.expression
}

func (c *DisjointCondition) BeforeEvaluate(ctx core.EvalContext) error {
	return ctx.BuildEvalParams(c.builtinParams)
}

func (c *DisjointCondition) Evaluate(ctx core.EvalContext) (bool, bool, error) {
	out, _, err := c.program.Eval(ctx.EvalParams())
	return Out(out, err)
}

func (c *DisjointCondition) Negate() (core.Condition, error) {
	n := NewIntersectCondition(c.field1, c.valueTypeStr1, c.field2, c.valueTypeStr2, c.cfg)
	if v := n.Validate(); !v.Valid {
		return nil, &InvalidNodeError{"got invalid node after negate"}
	}
	err := n.Compile()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (c *DisjointCondition) Compile() error {
	var builtinParamTypes map[string]*cel.Type = nil
	var builtinParams map[string]interface{} = nil
	left := c.field1
	b := false
	if c.valueType1 == core.ValueTypeList {
		var v []interface{}
		err := json.Unmarshal([]byte(c.field1), &v)
		if err != nil {
			return err
		}
		innerKey := builtin_inner_key_prefix + "list_001"
		left = builtin_var + "." + innerKey
		b = true
		builtinParams = mapOf(builtin_var, mapOf(innerKey, v))
	}
	right := c.field2
	if c.valueType2 == core.ValueTypeList {
		var v []interface{}
		err := json.Unmarshal([]byte(c.field2), &v)
		if err != nil {
			return err
		}
		innerKey := builtin_inner_key_prefix + "list_002"
		right = builtin_var + "." + innerKey
		b = true
		if builtinParams == nil {
			builtinParams = mapOf(builtin_var, mapOf(innerKey, v))
		} else {
			m := builtinParams[builtin_var].(map[string]interface{})
			m[innerKey] = v
			builtinParams[builtin_var] = m
		}
	}
	if len(builtinParams) > 0 {
		c.builtinParams = builtinParams
	}
	c.expression = fmt.Sprintf("!( %s.exists(x, %s.exists(y, y == x)) )", left, right)
	if b {
		builtinParamTypes = mapOfCelType(builtin_var, cel.MapType(cel.StringType, cel.DynType))
		return c.buildCELProgram(builtinParamTypes)
	}
	return c.buildCELProgram(nil)
}

func (c *DisjointCondition) generateName() string {
	return fmt.Sprintf("list(%s) and list(%s) are disjoint", c.field1, c.field2)
}

func (c *DisjointCondition) buildCELProgram(builtinVars map[string]*cel.Type) error {
	prg, err := BuildCELProgram(c.expression, builtinVars)
	if err != nil {
		return err
	}
	c.program = prg
	return nil
}
