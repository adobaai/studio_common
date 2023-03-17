package adapter

import (
	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
)

const casbinRuleTable = "casbin_rule"

type CasbinRule struct {
	ID    int64  `db:"id"`
	PType string `db:"p_type"`
	V0    string `db:"v0"`
	V1    string `db:"v1"`
	V2    string `db:"v2"`
	V3    string `db:"v3"`
	V4    string `db:"v4"`
	V5    string `db:"v5"`
}

var casbinRuleStruct = sqlbuilder.NewStruct(new(CasbinRule))

func ExistCasbinRule(db *sqlx.DB, rule *CasbinRule) (exist bool, err error) {
	sb := casbinRuleStruct.SelectFrom(casbinRuleTable)
	sb.Select(sb.As("COUNT(*)", "count"))
	sb.Where(rule.combineE(&sb.Cond)...)
	sqlStr, args := sb.Build()
	var num int
	err = db.Get(&num, sqlStr, args...)
	exist = num >= 1
	return
}

func ListCasbinRules(db *sqlx.DB) (list []*CasbinRule, err error) {
	sb := casbinRuleStruct.SelectFrom(casbinRuleTable)
	sqlStr, args := sb.Build()
	err = db.Select(&list, sqlStr, args...)
	return
}

func NewCasbinRule(db *sqlx.DB, rule *CasbinRule) error {
	ib := casbinRuleStruct.InsertInto(casbinRuleTable, rule)
	sqlStr, args := ib.Build()
	_, err := db.Exec(sqlStr, args...)
	return err
}

func DeleteCasbinRule(db *sqlx.DB, rule *CasbinRule) error {
	deb := casbinRuleStruct.DeleteFrom(casbinRuleTable)
	deb.Where(rule.combineE(&deb.Cond)...)
	sqlStr, args := deb.Build()
	_, err := db.Exec(sqlStr, args...)
	return err
}

func (rule *CasbinRule) combineE(c *sqlbuilder.Cond) (conditions []string) {
	conditions = append(conditions, c.E("p_type", rule.PType))
	if len(rule.V0) > 0 {
		conditions = append(conditions, c.E("v0", rule.V0))
	}
	if len(rule.V1) > 0 {
		conditions = append(conditions, c.E("v1", rule.V1))
	}
	if len(rule.V2) > 0 {
		conditions = append(conditions, c.E("v2", rule.V2))
	}
	if len(rule.V3) > 0 {
		conditions = append(conditions, c.E("v3", rule.V3))
	}
	if len(rule.V4) > 0 {
		conditions = append(conditions, c.E("v4", rule.V4))
	}
	if len(rule.V5) > 0 {
		conditions = append(conditions, c.E("v5", rule.V5))
	}
	return
}
