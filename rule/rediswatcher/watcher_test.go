package rediswatcher

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func initWatcherOptions(t *testing.T, ignore, noSub bool) (op *WatcherOptions) {
	testE, err := casbin.NewEnforcer("../examples/rbac_model.conf", "../examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	rds := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("127.0.0.1:6379")})
	op = &WatcherOptions{
		Rds:         rds,
		Channel:     "test.casbin",
		E:           testE,
		IgnoreSelf:  ignore,
		NoSubscribe: noSub,
		Log:         NewLogger(),
	}
	return
}

func initWatcher(t *testing.T, op *WatcherOptions) (w *Watcher) {
	iw, err := NewWatcher(op)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	w = iw.(*Watcher)
	_ = op.E.SetWatcher(w)
	return
}

func TestWatcher(t *testing.T) {
	op := initWatcherOptions(t, true, false)
	w := initWatcher(t, op)
	_ = w.SetUpdateCallback(func(s string) {
		t.Logf("test watcher %s", s)
	})
	_ = w.Update()
	time.Sleep(time.Second)
	w.Close()
}

func TestUpdate(t *testing.T) {
	op := initWatcherOptions(t, false, false)
	w := initWatcher(t, op)
	_ = w.SetUpdateCallback(func(s string) {
		m := new(MSG)
		if err := m.UnmarshalBinary([]byte(s)); err != nil {
			t.Error(err)
			return
		}
		if m.Method != "Update" {
			t.Errorf("Method should be Update instead of %v", m.Method)
		}
	})
	_ = w.Update()
	time.Sleep(time.Second)
	w.Close()
}

func TestUpdateForAddPolicy(t *testing.T) {
	op1 := initWatcherOptions(t, true, true)
	w1 := initWatcher(t, op1)

	op2 := initWatcherOptions(t, false, false)
	w2 := initWatcher(t, op2)

	time.Sleep(time.Second)
	if _, err := op1.E.AddPolicy("4", "/api/test", "GET"); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)

	if !reflect.DeepEqual(op1.E.GetPolicy(), op2.E.GetPolicy()) {
		t.Error("These two enforcers' policies should be equal")
		t.Log("e1.policies", op1.E.GetPolicy())
		t.Log("e2.policies", op2.E.GetPolicy())
	}
	w1.Close()
	w2.Close()
}

func TestUpdateForRemovePolicy(t *testing.T) {
	op1 := initWatcherOptions(t, true, true)
	w1 := initWatcher(t, op1)

	op2 := initWatcherOptions(t, false, false)
	w2 := initWatcher(t, op2)

	time.Sleep(time.Second)
	if _, err := op1.E.RemovePolicy("3", "/api/wallet", "GET"); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)

	if !reflect.DeepEqual(op1.E.GetPolicy(), op2.E.GetPolicy()) {
		t.Error("These two enforcers' policies should be equal")
		t.Log("e1.policies", op1.E.GetPolicy())
		t.Log("e2.policies", op2.E.GetPolicy())
	}
	w1.Close()
	w2.Close()
}

func TestUpdateForAddPolicies(t *testing.T) {
	op1 := initWatcherOptions(t, true, true)
	w1 := initWatcher(t, op1)

	op2 := initWatcherOptions(t, false, false)
	w2 := initWatcher(t, op2)

	time.Sleep(time.Second)
	rules := [][]string{
		{"4", "/api/test1", "GET"},
		{"5", "/api/test2", "GET"},
		{"6", "/api/test3", "GET"},
	}
	if _, err := op1.E.AddPolicies(rules); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)

	if !reflect.DeepEqual(op1.E.GetPolicy(), op2.E.GetPolicy()) {
		t.Error("These two enforcers' policies should be equal")
		t.Log("e1.policies", op1.E.GetPolicy())
		t.Log("e2.policies", op2.E.GetPolicy())
	}

	w1.Close()
	w2.Close()
}

func TestEnforceUpdatedForGroupPolicy(t *testing.T) {
	op1 := initWatcherOptions(t, true, true)
	w1 := initWatcher(t, op1)

	op2 := initWatcherOptions(t, false, false)
	w2 := initWatcher(t, op2)

	time.Sleep(time.Second)
	if _, err := op1.E.AddGroupingPolicy("tester", "1"); err != nil {
		t.Error(err)
	}
	t.Log("e1.group.policies", op1.E.GetGroupingPolicy())
	time.Sleep(time.Second * 2)

	res, err := op2.E.Enforce("tester", "/api/user", "GET")
	if err != nil {
		t.Error(err)
	}
	t.Log("e2.group.policies", op2.E.GetGroupingPolicy())
	assert.True(t, res)

	w1.Close()
	w2.Close()
}

func TestUpdateForRemoveFilteredPolicy(t *testing.T) {
	op1 := initWatcherOptions(t, true, true)
	w1 := initWatcher(t, op1)

	op2 := initWatcherOptions(t, false, false)
	w2 := initWatcher(t, op2)

	time.Sleep(time.Second)
	t.Log("before e1.polices", op1.E.GetPolicy())
	if _, err := op1.E.RemoveFilteredPolicy(0, "1", "/api/user", "DEL"); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second * 2)
	t.Log("after e2.polices", op2.E.GetPolicy())

	res := op2.E.HasPolicy("1", "/api/user", "DEL")
	assert.False(t, res)

	w1.Close()
	w2.Close()
}
