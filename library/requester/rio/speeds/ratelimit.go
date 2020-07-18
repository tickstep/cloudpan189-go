package speeds

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	RateLimit struct {
		MaxRate int64

		starTimestamp    int64
		count           int64
		interval        time.Duration
		ticker          *time.Ticker
		muChan          chan struct{}
		closeChan       chan struct{}
		backServiceOnce sync.Once
	}

	// AddCountFunc func() (count int64)
)

func NewRateLimit(maxRate int64) *RateLimit {
	return &RateLimit{
		MaxRate: maxRate,
	}
}

func (rl *RateLimit) SetInterval(i time.Duration) {
	if i <= 0 {
		i = 1 * time.Second
	}
	rl.interval = i
	if rl.ticker != nil {
		rl.ticker.Stop()
		rl.ticker = time.NewTicker(i)
	}
}

func (rl *RateLimit) Stop() {
	if rl.ticker != nil {
		rl.ticker.Stop()
	}
	if rl.closeChan != nil {
		close(rl.closeChan)
	}
	return
}

func (rl *RateLimit) resetChan() {
	if rl.muChan != nil {
		close(rl.muChan)
	}
	rl.muChan = make(chan struct{})
}

func (rl *RateLimit) backService() {
	if rl.interval <= 0 {
		rl.interval = 200 * time.Millisecond
	}
	rl.ticker = time.NewTicker(rl.interval)
	rl.closeChan = make(chan struct{})
	rl.resetChan()
	rl.starTimestamp = time.Now().UnixNano()
	go func() {
		for {
			select {
			case <-rl.ticker.C:
				if rl.rate() <= rl.MaxRate {
					rl.resetChan()
				}
			case <-rl.closeChan:
				return
			}
		}
	}()
}

func (rl *RateLimit) Add(count int64) {
	rl.backServiceOnce.Do(rl.backService)
	for {
		if rl.rate() >= rl.MaxRate { // 超出最大限额
			// 阻塞
			<-rl.muChan
			continue
		}
		atomic.AddInt64(&rl.count, count)
		if atomic.LoadInt64(&rl.count) < 0 {
			// reach the max value
			atomic.StoreInt64(&rl.count, 0)
			rl.starTimestamp = time.Now().Unix()
		}
		break
	}
}

func (rl *RateLimit) rate() int64 {
	timeElapseSecond := (time.Now().UnixNano() - rl.starTimestamp) / 1e9
	if timeElapseSecond <= 0 {
		timeElapseSecond = 1
	}
	return atomic.LoadInt64(&rl.count) / (timeElapseSecond)
}
