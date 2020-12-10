package prof

import (
	"bytes"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/threading"
)

type (
	Slot struct {
		lifecount int64
		lastcount int64
		lifecycle int64
		lastcycle int64
	}

	ProfileCenter struct {
		lock  sync.RWMutex
		slots map[string]*Slot
	}
)

const flushInterval = 5 * time.Minute

var (
	profileCenter = &ProfileCenter{
		slots: make(map[string]*Slot),
	}
	once sync.Once
)

func report(name string, duration time.Duration) {
	updated := func() bool {
		//如果slot已经存在，只需要读slots，更新其中一个slot，用读锁
		profileCenter.lock.RLock()
		defer profileCenter.lock.RUnlock()

		slot, ok := profileCenter.slots[name]
		if ok {
			atomic.AddInt64(&slot.lifecount, 1)
			atomic.AddInt64(&slot.lastcount, 1)
			atomic.AddInt64(&slot.lifecycle, int64(duration))
			atomic.AddInt64(&slot.lastcycle, int64(duration))
		}
		return ok
	}()

	if !updated {
		func() {
			//如果slot不存在，需要写slots，用写锁
			profileCenter.lock.Lock()
			defer profileCenter.lock.Unlock()

			profileCenter.slots[name] = &Slot{
				lifecount: 1,
				lastcount: 1,
				lifecycle: int64(duration),
				lastcycle: int64(duration),
			}
		}()
	}

	//单例启动一个goroutine去不断flush报告
	once.Do(flushRepeatly)
}

func flushRepeatly() {
	threading.GoSafe(func() {
		for {
			//每5分钟输出profile报告到日志文件stat.log
			time.Sleep(flushInterval)
			logx.Stat(generateReport())
		}
	})
}

func generateReport() string {
	var buffer bytes.Buffer
	buffer.WriteString("Profiling report\n")
	var data [][]string
	calcFn := func(total, count int64) string {
		if count == 0 {
			return "-"
		} else {
			return (time.Duration(total) / time.Duration(count)).String()
		}
	}

	func() {
		profileCenter.lock.Lock()
		defer profileCenter.lock.Unlock()

		for key, slot := range profileCenter.slots {
			data = append(data, []string{
				key,
				strconv.FormatInt(slot.lifecount, 10),
				calcFn(slot.lifecycle, slot.lifecount),
				strconv.FormatInt(slot.lastcount, 10),
				calcFn(slot.lastcycle, slot.lastcount),
			})

			// reset the data for last cycle
			slot.lastcount = 0
			slot.lastcycle = 0
		}
	}()

	table := tablewriter.NewWriter(&buffer)
	table.SetHeader([]string{"QUEUE", "LIFECOUNT", "LIFECYCLE", "LASTCOUNT", "LASTCYCLE"})
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()

	return buffer.String()
}
