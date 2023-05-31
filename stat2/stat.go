package stat2

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

// Stats 聚合列表
type Stats struct {
	delay time.Duration
	// interval time.Duration // 硬编码每秒统计和聚合

	statsMapMu sync.RWMutex
	statsMap   map[int64]*Bucket // 按秒粒度的时间戳 => Bucket
	sq         *StatQueue
}

func NewStats(delay time.Duration) *Stats {
	ss := &Stats{
		delay:    delay,
		statsMap: make(map[int64]*Bucket),
		sq:       new(StatQueue),
	}

	go ss.process()

	return ss
}

// 每秒一个篮子，收集此秒内的发送和收包情况.
type Bucket struct {
	Timestamp int64 // 按秒粒度的时间戳

	StatMu sync.Mutex
	Sent   map[uint64]int64 // 发送记录, key为seq
	Recv   map[uint64]int64 // 接收记录, key为seq
}

func NewBucket(ts int64) *Bucket {
	return &Bucket{
		Timestamp: ts,
		Sent:      make(map[uint64]int64),
		Recv:      make(map[uint64]int64),
	}
}

func (st *Bucket) WithLock(f func()) {
	st.StatMu.Lock()
	f()
	st.StatMu.Unlock()
}

// 输出此篮子的聚合信息
func (st *Bucket) PrintStat() {
	t := time.Unix(st.Timestamp, 0)

	sentTotal := len(st.Sent)
	recvTotal := 0
	latency := int64(0)

	for seq := range st.Sent {
		if _, ok := st.Recv[seq]; ok {
			recvTotal++
			latency += st.Recv[seq] - st.Sent[seq]
		}
	}

	if recvTotal > 0 {
		latency /= int64(recvTotal)
	}

	loss := 1 - float64(recvTotal)/float64(sentTotal)
	fmt.Printf("%s: sent=%d, recv=%d, loss rate: %.2f%%, latency=%d ns\n", t.Format("15:04:05"), sentTotal, recvTotal, loss*100, latency)

}

// AddSent 添加发送记录
func (ss *Stats) AddSent(seq uint64, ts int64) {
	key := ts / int64(time.Second)
	ss.statsMapMu.Lock()
	bucket := ss.statsMap[key] // 寻找篮子
	if bucket == nil {
		bucket = NewBucket(key)
		ss.statsMap[key] = bucket
		heap.Push(ss.sq, bucket)
	}
	ss.statsMapMu.Unlock()

	bucket.WithLock(func() {
		bucket.Sent[seq] = ts
	})
}

// AddRecv 添加接收记录
func (ss *Stats) AddRecv(seq uint64, ts int64) {
	key := ts / int64(time.Second)
	ss.statsMapMu.RLock()
	bucket := ss.statsMap[key]
	if bucket == nil {
		ss.statsMapMu.RUnlock()
		return
	}
	ss.statsMapMu.RUnlock()

	bucket.WithLock(func() {
		bucket.Recv[seq] = time.Now().UnixNano()
	})

}

func (ss *Stats) process() {
	ticker := time.NewTicker(time.Second)
	time.Sleep(ss.delay)
	for {
		st := heap.Pop(ss.sq)
		if st == nil {
			time.Sleep(ss.delay)
			continue
		}

		// 显示最久的聚合的数据
		bucket := st.(*Bucket)
		bucket.WithLock(func() {
			bucket.PrintStat()
		})

		<-ticker.C
	}
}

func (ss *Stats) Close() {
	// drain
	for {
		if bucket, ok := heap.Pop(ss.sq).(*Bucket); ok {
			bucket.WithLock(func() {
				bucket.PrintStat()
			})
		}
	}
}
