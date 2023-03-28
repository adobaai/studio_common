## Description
Studio的用户权限验证器生成、`rediswatcher`策略同步

## Watcher
考虑使用`redis`来保持多个`casbin`执行器实例之间的一致性。
可以直接借助`redis`的消息发布订阅机制，实现各个节点之间策略数据的增量同步

### Adopter

使用数据库，表名为`casbin_rule`

### WatcherEx interface

代码实现的是`WatcherEx`接口，与`watch`接口相比，可以区分更新的动作类型，如：更新策略，删除策略

```go
// WatcherEx is the strengthened Casbin watchers.
type WatcherEx interface {
	Watcher
	// UpdateForAddPolicy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.AddPolicy()
	UpdateForAddPolicy(sec, ptype string, params ...string) error
	// UpdateForRemovePolicy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.RemovePolicy()
	UpdateForRemovePolicy(sec, ptype string, params ...string) error
	// UpdateForRemoveFilteredPolicy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.RemoveFilteredNamedGroupingPolicy()
	UpdateForRemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error
	// UpdateForSavePolicy calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.RemoveFilteredNamedGroupingPolicy()
	UpdateForSavePolicy(model model.Model) error
	// UpdateForAddPolicies calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.AddPolicies()
	UpdateForAddPolicies(sec string, ptype string, rules ...[]string) error
	// UpdateForRemovePolicies calls the update callback of other instances to synchronize their policy.
	// It is called after Enforcer.RemovePolicies()
	UpdateForRemovePolicies(sec string, ptype string, rules ...[]string) error
}
```

### 实现机制
Redis的 [SUBSCRIBE](https://redis.io/docs/manual/pubsub/) 命令可以让客户端订阅任意数量的频道， 每当有新信息发送到被订阅的频道时， 信息就会被发送给所有订阅指定频道的客户端。
