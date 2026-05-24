package rules

import (
	"fmt"

	"github.com/franklee-labs/celero-go/condition"
	"github.com/franklee-labs/celero-go/core"
)

type ConditionCreator interface {
	Create(node ConditionNode) (core.Condition, error)
}

type ConditionCreateFunc func(node ConditionNode) (core.Condition, error)

func (f ConditionCreateFunc) Create(node ConditionNode) (core.Condition, error) {
	return f(node)
}

func StringValue(props map[string]interface{}, key string) (string, error) {
	val, has := props[key]
	if !has {
		return "", fmt.Errorf("no such key:%s", key)
	}
	if v, ok := val.(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("%s value is not string type", key)
}

func IntValue(props map[string]interface{}, key string) (int, error) {
	val, has := props[key]
	if !has {
		return 0, fmt.Errorf("no such key:%s", key)
	}
	if v, ok := val.(float64); ok {
		return int(v), nil
	}
	return 0, fmt.Errorf("%s value is not number type", key)
}

func BoolValue(props map[string]interface{}, key string) (bool, error) {
	val, has := props[key]
	if !has {
		return false, fmt.Errorf("no such key:%s", key)
	}
	if v, ok := val.(bool); ok {
		return v, nil
	}
	return false, fmt.Errorf("%s value is not bool type", key)
}

func BuildConditionConfig(conditionNode ConditionNode) (*core.ConditionConfig, error) {
	props := conditionNode.Properties()
	id := conditionNode.ID
	if id == "" {
		return nil, fmt.Errorf("condition node missing 'id'")
	}
	name := conditionNode.Name
	// if name == "" {
	// 	return nil, fmt.Errorf("condition node missing 'name'")
	// }
	priority, err := IntValue(props, "priority")
	if err != nil {
		return nil, err
	}
	cacheable := conditionNode.IsCacheable()
	ignoreAbsence := conditionNode.IsIgnoreAbsence()
	return core.NewConditionConfig(id, name).WithCacheable(cacheable).WithIgnoreAbsence(ignoreAbsence).WithPriority(priority), nil
}

func EqualConditionCreatefunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field, err := StringValue(props, "field")
	if err != nil {
		return nil, err
	}
	value, err := StringValue(props, "value")
	if err != nil {
		return nil, err
	}
	vt, _ := StringValue(props, "valueType")
	valueType := core.ValueTypeFromString(vt)
	if valueType == core.ValueTypeInvalid {
		return nil, fmt.Errorf("invalid value type")
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewEqualCondition(field, value, valueType, cfg), nil
}

func NotEqualConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field, err := StringValue(props, "field")
	if err != nil {
		return nil, err
	}
	value, err := StringValue(props, "value")
	if err != nil {
		return nil, err
	}
	vt, _ := StringValue(props, "valueType")
	valueType := core.ValueTypeFromString(vt)
	if valueType == core.ValueTypeInvalid {
		return nil, fmt.Errorf("invalid value type")
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewNotEqualCondition(field, value, valueType, cfg), nil
}

func CompareConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field, err := StringValue(props, "field")
	if err != nil {
		return nil, err
	}
	value, err := StringValue(props, "value")
	if err != nil {
		return nil, err
	}
	vt, _ := StringValue(props, "valueType")
	valueType := core.ValueTypeFromString(vt)
	if valueType == core.ValueTypeInvalid {
		return nil, fmt.Errorf("invalid value type")
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewCompareCondition(field, value, node.Sign, valueType, cfg), nil
}

func CELConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	expression, err := StringValue(props, "expression")
	if err != nil {
		return nil, err
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewCELCondition(expression, cfg), nil
}

func ExistsConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field, err := StringValue(props, "field")
	if err != nil {
		return nil, err
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewExistsCondition(field, cfg), nil
}

func AbsentConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field, err := StringValue(props, "field")
	if err != nil {
		return nil, err
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewAbsentCondition(field, cfg), nil
}

func InConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field, err := StringValue(props, "field")
	if err != nil {
		return nil, err
	}
	value, err := StringValue(props, "value")
	if err != nil {
		return nil, err
	}
	vt, _ := StringValue(props, "valueType")
	valueType := core.ValueTypeFromString(vt)
	if valueType == core.ValueTypeInvalid {
		return nil, fmt.Errorf("invalid value type")
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewInCondition(field, value, valueType, cfg), nil
}

func NotInConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field, err := StringValue(props, "field")
	if err != nil {
		return nil, err
	}
	value, err := StringValue(props, "value")
	if err != nil {
		return nil, err
	}
	vt, _ := StringValue(props, "valueType")
	valueType := core.ValueTypeFromString(vt)
	if valueType == core.ValueTypeInvalid {
		return nil, fmt.Errorf("invalid value type")
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewNotInCondition(field, value, valueType, cfg), nil
}

func RegexpConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field, err := StringValue(props, "field")
	if err != nil {
		return nil, err
	}
	regexp, err := StringValue(props, "regexp")
	if err != nil {
		return nil, err
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewRegexpCondition(field, regexp, cfg), nil
}

func IntersectConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field1, err := StringValue(props, "field1")
	if err != nil {
		return nil, err
	}
	valueType1, err := StringValue(props, "valueType1")
	if err != nil {
		return nil, err
	}
	field2, err := StringValue(props, "field2")
	if err != nil {
		return nil, err
	}
	valueType2, err := StringValue(props, "valueType2")
	if err != nil {
		return nil, err
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewIntersectCondition(field1, valueType1, field2, valueType2, cfg), nil
}

func DisjointConditionCreateFunc(node ConditionNode) (core.Condition, error) {
	props := node.Properties()
	field1, err := StringValue(props, "field1")
	if err != nil {
		return nil, err
	}
	valueType1, err := StringValue(props, "valueType1")
	if err != nil {
		return nil, err
	}
	field2, err := StringValue(props, "field2")
	if err != nil {
		return nil, err
	}
	valueType2, err := StringValue(props, "valueType2")
	if err != nil {
		return nil, err
	}
	cfg, err := BuildConditionConfig(node)
	if err != nil {
		return nil, err
	}
	return condition.NewDisjointCondition(field1, valueType1, field2, valueType2, cfg), nil
}
