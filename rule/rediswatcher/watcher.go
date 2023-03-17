package rediswatcher

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	rds "github.com/go-redis/redis"
)

type Watcher struct {
	mux      sync.Mutex
	pubSub   *rds.PubSub
	options  *WatcherOptions
	close    chan struct{}
	callback func(string)
}

func (w *Watcher) defaultUpdateCallback(e casbin.IEnforcer) (f func(string)) {
	l := w.options.Log
	f = func(msg string) {
		msgStruct := &MSG{}

		err := msgStruct.UnmarshalBinary([]byte(msg))
		if err != nil {
			l.Error(err)
			return
		}

		var res bool
		switch msgStruct.Method {
		case Update, UpdateForSavePolicy:
			res = true
		case UpdateForAddPolicy:
			res, err = e.SelfAddPolicy(msgStruct.Sec, msgStruct.Ptype, msgStruct.NewRule)
		case UpdateForAddPolicies:
			res, err = e.SelfAddPolicies(msgStruct.Sec, msgStruct.Ptype, msgStruct.NewRules)
		case UpdateForRemovePolicy:
			res, err = e.SelfRemovePolicy(msgStruct.Sec, msgStruct.Ptype, msgStruct.NewRule)
		case UpdateForRemoveFilteredPolicy:
			res, err = e.SelfRemoveFilteredPolicy(msgStruct.Sec, msgStruct.Ptype, msgStruct.FieldIndex, msgStruct.FieldValues...)
		case UpdateForRemovePolicies:
			res, err = e.SelfRemovePolicies(msgStruct.Sec, msgStruct.Ptype, msgStruct.NewRules)
		case UpdateForUpdatePolicy:
			res, err = e.SelfUpdatePolicy(msgStruct.Sec, msgStruct.Ptype, msgStruct.OldRule, msgStruct.NewRule)
		case UpdateForUpdatePolicies:
			res, err = e.SelfUpdatePolicies(msgStruct.Sec, msgStruct.Ptype, msgStruct.OldRules, msgStruct.NewRules)
		default:
			err = errors.New("unknown update type")
		}
		if err != nil {
			l.Errorf("callback err: %v", err)
		}
		if !res {
			l.Error("callback update policy failed")
		}
	}
	return
}

type MSG struct {
	Method      UpdateType
	ID          string
	Sec         string
	Ptype       string
	OldRule     []string
	OldRules    [][]string
	NewRule     []string
	NewRules    [][]string
	FieldIndex  int
	FieldValues []string
}

type UpdateType string

const (
	Update                        UpdateType = "Update"
	UpdateForAddPolicy            UpdateType = "UpdateForAddPolicy"
	UpdateForRemovePolicy         UpdateType = "UpdateForRemovePolicy"
	UpdateForRemoveFilteredPolicy UpdateType = "UpdateForRemoveFilteredPolicy"
	UpdateForSavePolicy           UpdateType = "UpdateForSavePolicy"
	UpdateForAddPolicies          UpdateType = "UpdateForAddPolicies"
	UpdateForRemovePolicies       UpdateType = "UpdateForRemovePolicies"
	UpdateForUpdatePolicy         UpdateType = "UpdateForUpdatePolicy"
	UpdateForUpdatePolicies       UpdateType = "UpdateForUpdatePolicies"
)

func (m *MSG) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary decodes the struct into a User
func (m *MSG) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, m); err != nil {
		return err
	}
	return nil
}

// NewWatcher creates a new Watcher to be used with a Casbin enforcer
// addr is a redis target string in the format "host:port"
// setters allows for inline WatcherOptions
func NewWatcher(op *WatcherOptions) (persist.Watcher, error) {
	w := &Watcher{
		close: make(chan struct{}),
	}
	if err := w.initOptions(op); err != nil {
		return nil, err
	}

	if !op.NoSubscribe {
		w.pubSub = op.Rds.Subscribe(op.Channel)
		w.subscribe()
	}

	return w, nil
}

func (w *Watcher) initOptions(op *WatcherOptions) (err error) {
	if err = initConfig(op); err != nil {
		return
	}
	w.options = op
	if op.OptionalUpdateCallback != nil {
		_ = w.SetUpdateCallback(op.OptionalUpdateCallback)
	} else {
		_ = w.SetUpdateCallback(func(s string) {
			w.defaultUpdateCallback(op.E)(s)
		})
	}
	return nil
}

// SetUpdateCallback sets the update callback function invoked by the watcher
// when the policy is updated. Defaults to Enforcer.LoadPolicy()
func (w *Watcher) SetUpdateCallback(callback func(string)) error {
	w.mux.Lock()
	w.callback = callback
	w.mux.Unlock()
	return nil
}

// Update publishes a message to all other instances telling them to
// invoke their update callback
// Enforcer.AddPolicy(), Enforcer.RemovePolicy(), etc.
func (w *Watcher) Update() error {
	f := func() error {
		msg := &MSG{
			Method: Update,
			ID:     w.options.LocalID,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, Update)
}

// UpdateForAddPolicy calls the update callback of other instances to synchronize their policy.
// It is called after Enforcer.AddPolicy()
func (w *Watcher) UpdateForAddPolicy(sec, ptype string, params ...string) error {
	f := func() error {
		msg := &MSG{
			Method:  UpdateForAddPolicy,
			ID:      w.options.LocalID,
			Sec:     sec,
			Ptype:   ptype,
			NewRule: params,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, UpdateForAddPolicy)
}

// UpdateForRemovePolicy calls the update callback of other instances to synchronize their policy.
// It is called after Enforcer.RemovePolicy()
func (w *Watcher) UpdateForRemovePolicy(sec, ptype string, params ...string) error {
	f := func() error {
		msg := &MSG{
			Method:  UpdateForRemovePolicy,
			ID:      w.options.LocalID,
			Sec:     sec,
			Ptype:   ptype,
			NewRule: params,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, UpdateForRemovePolicy)
}

// UpdateForRemoveFilteredPolicy calls the update callback of other instances to synchronize their policy.
// It is called after Enforcer.RemoveFilteredNamedGroupingPolicy()
func (w *Watcher) UpdateForRemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	f := func() error {
		msg := &MSG{
			Method:      UpdateForRemoveFilteredPolicy,
			ID:          w.options.LocalID,
			Sec:         sec,
			Ptype:       ptype,
			FieldIndex:  fieldIndex,
			FieldValues: fieldValues,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, UpdateForRemoveFilteredPolicy)
}

// UpdateForSavePolicy calls the update callback of other instances to synchronize their policy.
// It is called after Enforcer.RemoveFilteredNamedGroupingPolicy()
func (w *Watcher) UpdateForSavePolicy(_ model.Model) error {
	f := func() error {
		msg := &MSG{
			Method: UpdateForSavePolicy,
			ID:     w.options.LocalID,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, UpdateForSavePolicy)
}

// UpdateForAddPolicies calls the update callback of other instances to synchronize their policies in batch.
// It is called after Enforcer.AddPolicies()
func (w *Watcher) UpdateForAddPolicies(sec string, ptype string, rules ...[]string) error {
	f := func() error {
		msg := &MSG{
			Method:   UpdateForAddPolicies,
			ID:       w.options.LocalID,
			Sec:      sec,
			Ptype:    ptype,
			NewRules: rules,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, UpdateForAddPolicies)
}

// UpdateForRemovePolicies calls the update callback of other instances to synchronize their policies in batch.
// It is called after Enforcer.RemovePolicies()
func (w *Watcher) UpdateForRemovePolicies(sec string, ptype string, rules ...[]string) error {
	f := func() error {
		msg := &MSG{
			Method:   UpdateForRemovePolicies,
			ID:       w.options.LocalID,
			Sec:      sec,
			Ptype:    ptype,
			NewRules: rules,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, UpdateForRemovePolicies)
}

// UpdateForUpdatePolicy calls the update callback of other instances to synchronize their policy.
// It is called after Enforcer.UpdatePolicy()
func (w *Watcher) UpdateForUpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	f := func() error {
		msg := &MSG{
			Method:  UpdateForUpdatePolicy,
			ID:      w.options.LocalID,
			Sec:     sec,
			Ptype:   ptype,
			OldRule: oldRule,
			NewRule: newRule,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, UpdateForUpdatePolicy)
}

// UpdateForUpdatePolicies calls the update callback of other instances to synchronize their policy.
// It is called after Enforcer.UpdatePolicies()
func (w *Watcher) UpdateForUpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	f := func() error {
		msg := &MSG{
			Method:   UpdateForUpdatePolicies,
			ID:       w.options.LocalID,
			Sec:      sec,
			Ptype:    ptype,
			OldRules: oldRules,
			NewRules: newRules,
		}
		return w.publish(msg)
	}
	return w.logRecord(f, UpdateForUpdatePolicies)
}

func (w *Watcher) Close() {
	w.mux.Lock()
	defer w.mux.Unlock()
	close(w.close)
	w.options.Rds.Publish(w.options.Channel, "close")
}

func (w *Watcher) logRecord(f func() error, t UpdateType) (err error) {
	l := w.options.Log
	if err = f(); err != nil {
		l.Errorf("[%s] err: %v", t, err)
	}
	return
}

func (w *Watcher) publish(msg *MSG) error {
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.options.Rds.Publish(w.options.Channel, msg).Err()
}

func (w *Watcher) subscribe() {
	var (
		sub = w.options.Rds.Subscribe(w.options.Channel)
		wg  = sync.WaitGroup{}
		l   = w.options.Log
	)

	wg.Add(1)
	go func() {
		defer func() {
			if err := sub.Close(); err != nil {
				l.Errorf("sub closed err: %v", err)
			}
			if err := w.pubSub.Close(); err != nil {
				l.Errorf("pubsub closed err: %v", err)
			}
		}()
		ch := sub.Channel()
		wg.Done()
		for msg := range ch {
			select {
			case <-w.close:
				return
			default:
			}
			data := msg.Payload
			if data == "close" {
				return
			}
			l.Infof("received message from channel %s", data)
			m := new(MSG)
			if err := m.UnmarshalBinary([]byte(data)); err != nil {
				l.Errorf("Failed to parse message: %s with error: %v\n", data, err)
			} else {
				isSelf := m.ID == w.options.LocalID
				if !(w.options.IgnoreSelf && isSelf) {
					w.callback(data)
				}
			}
		}
	}()
	wg.Wait()
}
