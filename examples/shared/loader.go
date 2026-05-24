package shared

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/franklee-labs/celero-go/engine"
)

//go:embed data/coupon-rules.json
var couponRulesData []byte

//go:embed data/coupon-rules-ignore-absence.json
var couponRulesIgnoreAbsenceData []byte

type ruleConfig struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Rule        json.RawMessage `json:"rule"`
}

func LoadCouponRules() ([]*engine.CeleroRule, error) {
	return loadRules(couponRulesData)
}

func LoadCouponRulesIgnoreAbsence() ([]*engine.CeleroRule, error) {
	return loadRules(couponRulesIgnoreAbsenceData)
}

func loadRules(data []byte) ([]*engine.CeleroRule, error) {
	var configs []ruleConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}
	rules := make([]*engine.CeleroRule, 0, len(configs))
	for _, cfg := range configs {
		rb, err := engine.FromJSON(cfg.ID, string(cfg.Rule))
		if err != nil {
			return nil, fmt.Errorf("parse rule %q: %w", cfg.ID, err)
		}
		cr, err := rb.Name(cfg.Name).Description(cfg.Description).Build()
		if err != nil {
			return nil, fmt.Errorf("build rule %q: %w", cfg.ID, err)
		}
		rules = append(rules, cr)
	}
	return rules, nil
}
