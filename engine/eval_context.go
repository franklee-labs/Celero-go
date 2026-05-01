package engine

import (
	"fmt"
	"strings"

	"github.com/franklee-labs/celero-go/core"
)

type EvalContext struct {
	ruleContext                    *RuleContext
	params                         map[string]interface{}
	evalParams                     map[string]interface{}
	enbleMissingState              bool
	enableConditionResultCacheable bool
	enableReport                   bool
	conditionEvalResult            map[string]core.EvalResult
}

func (c *EvalContext) Params() map[string]interface{} {
	return c.ruleContext.GetParams()
}

func (c *EvalContext) RuleContext() core.RuleContext {
	return c.ruleContext
}

func (c *EvalContext) GetAttributes() map[string]interface{} {
	return c.ruleContext.GetAttributes()
}

func (c *EvalContext) GetAttribute(key string) (interface{}, bool) {
	return c.ruleContext.GetAttribute(key)
}

func (c *EvalContext) SetAttribute(key string, value interface{}) {
	c.ruleContext.SetAttribute(key, value)
}

func (c *EvalContext) IsConditionResultCacheEnabled() bool {
	return c.enableConditionResultCacheable
}

func (c *EvalContext) IsAbsenceEnabled() bool {
	return c.enbleMissingState
}

func (c *EvalContext) IsReportEnabled() bool {
	return c.ruleContext.IsReportEnabled()
}

// BuildEvalParams merges runtime params with condition-specific builtin params
// into evalParams. Called by BeforeEvaluate so Evaluate can read a single map.
func (c *EvalContext) BuildEvalParams(builtinParams map[string]interface{}) error {
	params := c.ruleContext.GetParams()
	merged := make(map[string]interface{}, len(params)+len(builtinParams))
	for k, v := range params {
		merged[k] = v
	}
	for k, v := range builtinParams {
		if !strings.HasPrefix(k, "_") {
			return fmt.Errorf("builtin keys must start with _ [key=%s]", k)
		}
		merged[k] = v
	}
	c.evalParams = merged
	return nil
}

func (c *EvalContext) EvalParams() map[string]interface{} {
	return c.evalParams
}

func (c *EvalContext) SetConditionEvalResult(conditionID string, result core.EvalResult) {
	c.conditionEvalResult[conditionID] = result
}

func (c *EvalContext) GetConditionEvalResult(conditionID string) (core.EvalResult, bool) {
	result, exists := c.conditionEvalResult[conditionID]
	return result, exists
}

type Builder struct {
	ruleContext              *RuleContext
	enableMissingState       bool
	conditionResultCacheable bool
	enableReport             bool
}

func NewBuilder(rc *RuleContext) *Builder {
	return &Builder{
		ruleContext:              rc,
		enableMissingState:       false,
		conditionResultCacheable: false,
		enableReport:             false,
	}
}

func (b *Builder) SetEnableMissingState(enable bool) *Builder {
	b.enableMissingState = enable
	return b
}

func (b *Builder) SetConditionResultCacheable(cacheable bool) *Builder {
	b.conditionResultCacheable = cacheable
	return b
}

func (b *Builder) EnableReport(enable bool) *Builder {
	b.enableReport = enable
	return b
}

func (b *Builder) Build() (*EvalContext, error) {
	return &EvalContext{
		ruleContext:                    b.ruleContext,
		params:                         b.ruleContext.params,
		enbleMissingState:              b.enableMissingState,
		enableConditionResultCacheable: b.conditionResultCacheable,
		enableReport:                   b.enableReport,
		conditionEvalResult:            make(map[string]core.EvalResult),
	}, nil
}
