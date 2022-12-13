package stat

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type AggrStat struct {
	statsMu sync.RWMutex
	stats   map[int64][]*Stat
}

func NewAggrStat() *AggrStat {
	return &AggrStat{
		stats: make(map[int64][]*Stat),
	}
}

func (as *AggrStat) addStat(st *Stat) {
	as.statsMu.Lock()
	as.stats[st.Timestamp] = append(as.stats[st.Timestamp], st)
	if len(as.stats) == 5 {
		min := int64(math.MaxInt64)
		for ts := range as.stats {
			if ts < min {
				min = ts
			}
		}
		ss := as.stats[min]
		delete(as.stats, min)
		as.statsMu.Unlock()
		as.printStats(min, ss)

	} else {
		as.statsMu.Unlock()
	}

}

func (as *AggrStat) printStats(ts int64, ss []*Stat) {
	sentTotal := 0
	recvTotal := 0
	latency := int64(0)

	for _, st := range ss {
		sentTotal += len(st.Sent)
		for seq := range st.Sent {
			if _, ok := st.Recv[seq]; ok {
				recvTotal++
				latency += st.Recv[seq] - st.Sent[seq]
			}
		}
	}

	if recvTotal > 0 {
		latency /= int64(recvTotal)
	}

	loss := 1 - float64(recvTotal)/float64(sentTotal)
	fmt.Printf("%s: sent=%d, recv=%d, loss rate: %.3f%%, latency=%d ns\n", time.Unix(ts, 0).Format("15:04:05"), sentTotal, recvTotal, loss*100, latency)

}
