package snowflake

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"sync"
	"time"
)

type ID struct {
	sync.Mutex
	timestamp int64 // 单位ms
	sequence  int64 // 序列号
	machineID int64 // 机器id
}

const (
	epoch         int64 = 1640966400000 // 起始时间: 2022-01-01 00:00:00
	machineIDBits uint8 = 5
	sequenceBits  uint8 = 12                        // 序列号所占的位数
	maxSequence   int64 = -1 ^ (-1 << sequenceBits) // 支持的最大序列号数量
	maxMachineId  int64 = -1 ^ (-1 << machineIDBits)
	keep32Bits    int64 = -1 ^ (-1 << 32) + 1 // %b: 1 0000 0000 0000 0000 0000 0000 0000 0000
)

func NewIDGenerator(mid int64) *ID {
	id := new(ID)
	if mid > maxMachineId || mid < 1 {
		num, _ := rand.Int(rand.Reader, big.NewInt(maxMachineId+1))
		mid = num.Int64()
	}
	id.machineID = mid
	return id
}

// NextUID User ID
func (id *ID) NextUID() uint32 {
	id.Lock()
	defer id.Unlock()

	now := time.Now().UnixMilli()
	if id.timestamp == now {
		// 如果当前序列超出`maxSequence`长度，则需要等待下一毫秒
		// 下一毫秒将设置sequence: 0
		id.sequence = (id.sequence + 1) & maxSequence
		if id.sequence == 0 {
			for now <= id.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		id.sequence = 0
	}
	id.timestamp = now

	// 1. 得到当前时间与预设的起始时间（epoch）之间的时间差：T1(ms)；
	// 2. 将T1左移`(machineIDBits + sequenceBits)`位，保留足够的空间给机器ID和序列号使用；
	// 3. 将机器ID移动到合适位置合并；
	// 4. 合并序列号。
	//
	// 这样，生成的唯一标识符就能够在高位正确地包含时间戳，中间位包含节点ID，低位包含序列号。
	t := ((id.timestamp - epoch) << (machineIDBits + sequenceBits)) |
		(id.machineID << sequenceBits) |
		id.sequence

	// 如果t超出uint32范围，则进行截断处理
	return uint32(t | keep32Bits)
}

// NextEID Enterprise ID
func (id *ID) NextEID(sequence int64) uint32 {
	now := time.Now().UnixMilli()

	t := ((now - epoch) << (machineIDBits + sequenceBits)) |
		(id.machineID << (sequenceBits)) |
		sequence

	num := uint32(t | keep32Bits)
	numStr := strconv.FormatUint(uint64(num), 10)
	if len(numStr) > 4 {
		numStr = "888" + numStr[4:]
	} else {
		numStr = "888" + numStr
	}

	eid, _ := strconv.ParseUint(numStr, 10, 32)

	return uint32(eid)
}
