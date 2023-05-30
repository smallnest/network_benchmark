package stat

import (
	"fmt"
	"sync"
	"time"
)

type Stats struct {
	key string

	statsMapMu sync.RWMutex
	statsMap   map[int64]*Stat

	stats []*Stat

	ch       chan *Stat
	aggrStat *AggrStat
}

func NewStats(key string, aggrStat *AggrStat) *Stats {
	ss := &Stats{
		key:      key,
		statsMap: make(map[int64]*Stat),
		ch:       make(chan *Stat, 102400),
		aggrStat: aggrStat,
	}

	go ss.process()

	return ss
}

type Stat struct {
	Timestamp int64 // 按秒粒度的时间戳

	StatMu sync.Mutex
	Sent   map[uint64]int64 // 发送记录, key为seq
	Recv   map[uint64]int64 // 接收记录, key为seq
}

func NewStat(ts int64) *Stat {
	return &Stat{
		Timestamp: ts,
		Sent:      make(map[uint64]int64),
		Recv:      make(map[uint64]int64),
	}
}

// AddSent 添加发送记录
func (ss *Stats) AddSent(seq uint64, ts int64) {
	key := ts / int64(time.Second)
	ss.statsMapMu.Lock()
	st := ss.statsMap[key] // 寻找篮子
	if st == nil {
		st = NewStat(key)
		ss.statsMap[key] = st
		select {
		case ss.ch <- st:
		default:
		}
	}
	ss.statsMapMu.Unlock()

	st.StatMu.Lock()
	st.Sent[seq] = ts
	st.StatMu.Unlock()

}

// AddRecv 添加接收记录
func (ss *Stats) AddRecv(seq uint64, ts int64) {
	key := ts / int64(time.Second)
	ss.statsMapMu.RLock()
	st := ss.statsMap[key]
	if st == nil {
		ss.statsMapMu.RUnlock()
		return
	}
	ss.statsMapMu.RUnlock()

	st.StatMu.Lock()
	st.Recv[seq] = time.Now().UnixNano()
	st.StatMu.Unlock()

}

func (ss *Stats) process() {
	for st := range ss.ch {
		ss.stats = append(ss.stats, st)
		for len(ss.stats) > 20 {
			st := ss.stats[0]
			ss.stats = ss.stats[1:]
			ss.statsMapMu.Lock()
			delete(ss.statsMap, st.Timestamp)
			ss.statsMapMu.Unlock()
			ss.aggrStat.addStat(st)
			// printStat(ss.key, st)
		}
	}
}

func printStat(key string, st *Stat) {
	t := time.Unix(0, st.Timestamp)

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
	fmt.Printf("%s:key:%s, sent=%d, recv=%d, loss rate: %.2f%%, latency=%d ns\n", t.Format("15:04:05"), key, sentTotal, recvTotal, loss*100, latency)

}

func (ss *Stats) Close() {
	// drain
	for _, st := range ss.stats {
		st.StatMu.Lock()
		printStat(ss.key, st)
		st.StatMu.Unlock()
	}
}
