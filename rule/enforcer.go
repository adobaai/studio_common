package rule

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/jmoiron/sqlx"

	"studio_common/rule/adapter"
)

const modelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && (keyMatch5(r.obj, p.obj) || keyMatch3(r.obj, p.obj)) && regexMatch(r.act, p.act)
`

func NewEnforcer(db *sqlx.DB) (*casbin.Enforcer, error) {
	m, _ := model.NewModelFromString(modelText)
	a := adapter.NewAdapter(db)
	return casbin.NewEnforcer(m, a)
}
