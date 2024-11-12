package lock

import (
	"fmt"
	"sync"
	"time"
)

func Entry() {

	var (
		kvs = map[string]string{"aa": "bb"}
		i   = 0
		m   sync.RWMutex
	)
	fmt.Println(kvs)
	for {
		time.Sleep(time.Millisecond * 100)
		go func() {
			m.Lock()
			kvs[fmt.Sprintf("%s%d", "cc", i)] = fmt.Sprintf("%s%d", "dd", i)
			m.Unlock()
		}()
		go func() {
			m.RLock()
			fmt.Println(kvs)
			m.RUnlock()
		}()
		i++
	}
}
