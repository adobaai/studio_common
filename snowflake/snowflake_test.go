package snowflake

import (
	"sync"
	"testing"
	"time"
)

func Test_TwoMachinesGenerateUID(t *testing.T) {
	mid := make(map[uint32]struct{})
	start := time.Now().UnixMilli()
	done := make(chan struct{})
	ch := make(chan uint32, 1000)

	go func() {
		for v := range ch {
			_, exist := mid[v]
			if exist {
				t.Errorf("generate repeated id: %d", v)
				break
			}
			mid[v] = struct{}{}
		}
		done <- struct{}{}
	}()

	wg := sync.WaitGroup{}
	num := 2
	wg.Add(num)
	for g := 1; g <= num; g++ {
		go func(g int) {
			id := NewUidGenerator(int64(g))
			for i := 0; i < 100; i++ {
				ch <- id.NextUID()
			}
			wg.Done()
		}(g)
	}
	wg.Wait()

	close(ch)
	<-done

	t.Logf("speed milli seconds: %d ms", time.Now().UnixMilli()-start)
}

func Test_10000UidGenerate(t *testing.T) {
	start := time.Now().UnixMilli()
	id := NewUidGenerator(0)
	mid := make(map[uint32]struct{})

	for i := 0; i < 10000; i++ {
		v := id.NextUID()
		_, exist := mid[v]
		if exist {
			t.Errorf("generate repeated id: %d", v)
			return
		}
		t.Logf("generate id : %d", v)
	}

	t.Logf("speed milli seconds: %d ms ", time.Now().UnixMilli()-start)
}
