package snowflake

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	HundredK = 1000 * 100
	M        = 10000 * 100
)

func Test_TwoMachinesGenerate100KUID(t *testing.T) {
	mid := make(map[uint32]struct{})
	start := time.Now().UnixMilli()
	done := make(chan struct{})
	ch := make(chan uint32, HundredK)

	fr := NewFiler()
	defer fr.close()

	go func() {
		for v := range ch {
			_, exist := mid[v]
			if exist {
				t.Errorf("generated the repeated id: %d", v)
				break
			}
			fr.write(v)
			mid[v] = struct{}{}
		}
		done <- struct{}{}
	}()

	wg := sync.WaitGroup{}
	num := 2
	wg.Add(num)

	for g := 1; g <= num; g++ {
		go func(g int) {
			id := NewIDGenerator(int64(g))
			for i := 0; i < HundredK; i++ {
				ch <- id.NextUID()
			}
			wg.Done()
		}(g)
	}
	wg.Wait()

	close(ch)
	<-done

	fr.flush()
	t.Logf("speed milli seconds: %d ms", time.Now().UnixMilli()-start)
}

func Test_Generates1MUID(t *testing.T) {
	start := time.Now().UnixMilli()
	id := NewIDGenerator(0)
	mid := make(map[uint32]struct{})

	fr := NewFiler()
	defer fr.close()

	for i := 0; i < M; i++ {
		v := id.NextUID()
		_, exist := mid[v]
		if exist {
			t.Errorf("generated the %dth repeated id: %d", i, v)
			return
		}
		mid[v] = struct{}{}
		fr.write(v)
	}

	fr.flush()
	t.Logf("speed milli seconds: %d ms ", time.Now().UnixMilli()-start)
}

func Test_Generates4095EID(t *testing.T) {
	var (
		id  = NewIDGenerator(0)
		mid = make(map[uint32]struct{})
		fr  = NewFiler()
	)
	defer fr.close()

	var i int64 = 0
	for ; i <= maxSequence; i++ {
		v := id.NextEID(i)
		_, exist := mid[v]
		if exist {
			t.Errorf("generate the %dth repeated id: %d", i, v)
			return
		}
		mid[v] = struct{}{}
		fr.write(v)
	}
	fr.flush()
}

type Filer struct {
	f *os.File
	w *bufio.Writer
}

func NewFiler() (fr *Filer) {
	fName := "id.txt"
	_, err := os.Stat(fName)
	var f *os.File
	if os.IsNotExist(err) {
		f, _ = os.Create(fName)
	} else {
		f, _ = os.OpenFile(fName, os.O_WRONLY|os.O_APPEND, 0666)
	}
	_ = f.Truncate(0)
	fr = &Filer{
		f: f,
		w: bufio.NewWriter(f),
	}
	return
}

func (fr *Filer) write(uid uint32) {
	_, err := fr.w.WriteString(fmt.Sprintf("%d\n", uid))
	if err != nil {
		log.Fatalf("write uid err: %v", err)
	}
}

func (fr *Filer) close() {
	_ = fr.f.Close()
}

func (fr *Filer) flush() {
	_ = fr.w.Flush()
}
