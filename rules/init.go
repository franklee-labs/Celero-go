package rules

import (
	"fmt"
	"slices"
	"strings"
)

var globalConditionCreator = make(map[string]ConditionCreator, 0)
var internals = []string{"AND", "OR", "NOT"}

var ruleConditionCreator = make(map[string]map[string]ConditionCreator, 0)

func init() {
	RegisterConditionCreator("EQ", ConditionCreateFunc(EqualConditionCreatefunc))
	RegisterConditionCreator("NEQ", ConditionCreateFunc(NotEqualConditionCreateFunc))
	RegisterConditionCreator("LT", ConditionCreateFunc(CompareConditionCreateFunc))
	RegisterConditionCreator("LTE", ConditionCreateFunc(CompareConditionCreateFunc))
	RegisterConditionCreator("GT", ConditionCreateFunc(CompareConditionCreateFunc))
	RegisterConditionCreator("GTE", ConditionCreateFunc(CompareConditionCreateFunc))
	RegisterConditionCreator("IN", ConditionCreateFunc(InConditionCreateFunc))
	RegisterConditionCreator("NIN", ConditionCreateFunc(NotInConditionCreateFunc))
	RegisterConditionCreator("REGEXP", ConditionCreateFunc(RegexpConditionCreateFunc))
	RegisterConditionCreator("CEL", ConditionCreateFunc(CELConditionCreateFunc))
	RegisterConditionCreator("EXISTS", ConditionCreateFunc(ExistsConditionCreateFunc))
	RegisterConditionCreator("ABSENT", ConditionCreateFunc(AbsentConditionCreateFunc))
	RegisterConditionCreator("INTERSECT", ConditionCreateFunc(IntersectConditionCreateFunc))
	RegisterConditionCreator("DISJOINT", ConditionCreateFunc(DisjointConditionCreateFunc))

	for k := range globalConditionCreator {
		internals = append(internals, k)
	}

}

func RegisterConditionCreator(sign string, creator ConditionCreator) error {
	if sign == "" {
		return fmt.Errorf("sign can not be empty")
	}
	sign = strings.ToUpper(sign)
	if slices.Contains(internals, sign) {
		return fmt.Errorf("sign must not be [%s]", sign)
	}
	if _, has := globalConditionCreator[sign]; has {
		return fmt.Errorf("duplicated condition creator sign. sign=[%s]", sign)
	}
	globalConditionCreator[sign] = creator
	return nil
}

func RegisterRuleConditionCreator(ruleID, sign string, creator ConditionCreator) error {
	if sign == "" {
		return fmt.Errorf("sign can not be empty")
	}
	sign = strings.ToUpper(sign)
	if slices.Contains(internals, sign) {
		return fmt.Errorf("sign must not be [%s]", sign)
	}
	v, has := ruleConditionCreator[ruleID]
	if !has {
		v = make(map[string]ConditionCreator, 0)
		v[sign] = creator
		ruleConditionCreator[ruleID] = v
		return nil
	}
	if _, has := v[sign]; has {
		return fmt.Errorf("duplicated condition creator sign. sign=[%s]", sign)
	}
	v[sign] = creator
	ruleConditionCreator[ruleID] = v
	return nil
}

func GetConditionCreator(ruleID, sign string) (ConditionCreator, error) {
	if sign == "" {
		return nil, fmt.Errorf("sign must not be empty")
	}
	sign = strings.ToUpper(sign)
	if v, has := ruleConditionCreator[ruleID]; has {
		if f, has := v[sign]; has {
			return f, nil
		}
	}
	if v, has := globalConditionCreator[sign]; has {
		return v, nil
	}
	return nil, fmt.Errorf("no condition creator for sign[%s]", sign)
}
