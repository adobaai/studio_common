package snowflake

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"
)

type uid struct {
	sync.Mutex
	timestamp int64  // 单位ms
	sequence  uint32 // 序列号
	machineID int64  // 机器id
}

const (
	epoch         int64  = 1640966400000 // 起始时间: 2022-01-01 00:00:00
	machineIDBits uint8  = 8
	sequenceBits  uint8  = 24                        // 序列所占的位数
	maxSequence   uint32 = -1 ^ (-1 << sequenceBits) // 支持的最大序列id数量
	maxMachineId  int64  = -1 ^ (-1 << machineIDBits)
	keep32Bits    int64  = -1 ^ (-1 << 32) + 1
)

func NewUidGenerator(mid int64) *uid {
	id := new(uid)
	if mid > maxMachineId || mid < 1 {
		num, _ := rand.Int(rand.Reader, big.NewInt(maxMachineId+1))
		mid = num.Int64()
	}
	id.machineID = mid
	return id
}

func (u *uid) NextUID() (uid uint32) {
	u.Lock()
	defer u.Unlock()

	now := time.Now().UnixMilli()
	if u.timestamp == now {
		// 如果当前序列超出24bits长度，则需要等待下一毫秒
		// 下一毫秒将设置sequence:0
		u.sequence = (u.sequence + 1) & maxSequence
		if u.sequence == 0 {
			for now <= u.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		u.sequence = 0
	}
	u.timestamp = now
	t := (u.timestamp - epoch) << sequenceBits
	t = t | u.machineID<<machineIDBits | int64(u.sequence)
	return uint32(t | keep32Bits)
}
