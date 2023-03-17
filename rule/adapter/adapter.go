package adapter

import (
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/jmoiron/sqlx"
)

type Adapter struct {
	db *sqlx.DB
}

func NewAdapter(db *sqlx.DB) *Adapter {
	return &Adapter{db}
}

func loadPolicyLine(line *CasbinRule, model model.Model) {
	lineText := line.PType
	if line.V0 != "" {
		lineText += ", " + line.V0
	}
	if line.V1 != "" {
		lineText += ", " + line.V1
	}
	if line.V2 != "" {
		lineText += ", " + line.V2
	}
	if line.V3 != "" {
		lineText += ", " + line.V3
	}
	if line.V4 != "" {
		lineText += ", " + line.V4
	}
	if line.V5 != "" {
		lineText += ", " + line.V5
	}
	persist.LoadPolicyLine(lineText, model)
}

func savePolicyLine(ptype string, rule []string) (line *CasbinRule) {
	line = new(CasbinRule)
	line.PType = ptype
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}
	return line
}

// LoadPolicy loads policy from database.
func (a *Adapter) LoadPolicy(model model.Model) error {
	lines, err := ListCasbinRules(a.db)
	if err != nil {
		return err
	}
	for _, line := range lines {
		loadPolicyLine(line, model)
	}
	return nil
}

// SavePolicy saves policy to database.
func (a *Adapter) SavePolicy(model model.Model) (err error) {
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			if err = NewCasbinRule(a.db, line); err != nil {
				return
			}
		}
	}
	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			if err = NewCasbinRule(a.db, line); err != nil {
				return
			}
		}
	}
	return
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(_ string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)
	exist, err := ExistCasbinRule(a.db, line)
	if err != nil {
		return err
	}
	if !exist {
		return NewCasbinRule(a.db, line)
	}
	return nil
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(_ string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)
	return DeleteCasbinRule(a.db, line)
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(_ string, ptype string, fieldIndex int, fieldValues ...string) error {
	line := &CasbinRule{}
	line.PType = ptype
	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		line.V5 = fieldValues[5-fieldIndex]
	}
	return DeleteCasbinRule(a.db, line)
}
