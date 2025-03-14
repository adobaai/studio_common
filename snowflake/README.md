# 生成Uid

## 为什么使用算法生成

### 数据库自增主键

如果是使用 mysql 数据库，那么通过设置主键为 auto_increment 是最容易实现单调递增的唯一ID 的方法，并且它也方便排序和索引。
但是缺点也很明显，由于过度依赖数据库，那么受限于数据库的性能会导致并发性并不高。

### Redis

在 Redis 中有两个命令 Incr、IncrBy ,因为Redis是单线程的所以通过这两个指令可以能保证原子性从而达到生成唯一值的目标，并且性能也很好。

但是在 Redis 中，即使有 AOF 和 RDB ，但是依然会存在数据丢失，有可能会造成ID重复。

## Snowflake

大多数使用`Snowflake`算法生成的uid多是64位，是为了兼顾到并发量；但是咱们系统的用户量，没必要使用64位的类型id，而且还占更多的存储空间，
这里我修改了算法，只会生成32位的id。

> 💡 经过测试，模拟两台单机在同一时刻内生成uid，1毫秒内可以生成200K个不重复的uid，每台100K个；单台机器的测试结果是生成1M个不重复的uid。
> 满足目前业务需求。

### 特点：
1. 能满足高并发分布式系统环境下ID不重复
2. 基于时间戳，可以保证基本有序递增
3. 不依赖于第三方的库或者中间件

### 如何提高UID的唯一性
1. 增加序列号的位数
2. 减少时间戳的粒度
3. 增加机器ID的位数

### 实现原理：

需要注意的是，由于-1在二进制上表示为

`11111111 11111111 11111111 11111111 11111111 11111111 11111111 11111111`

所以要知道32bits的最大值，可以讲-1向左移动32位，得到:

`11111111 11111111 11111111 11111111 00000000 00000000 00000000 00000000`

再和-1进行^异或运算:

`00000000 00000000 00000000 00000000 11111111 11111111 11111111 11111111`

### 实现步骤

1. 获取当前的毫秒时间戳；
2. 用当前的毫秒时间戳和上次保存的时间戳进行比较；
    1. 如果和上次保存的时间戳相等，那么对序列号 sequence 加一；
    2. 如果不相等，那么直接设置 sequence 为 0 即可；
3. 然后通过或运算拼接雪花算法需要返回的 `uint32` 返回值。

