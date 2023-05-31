package stat2

// StatQueue stat的堆
type StatQueue []*Bucket

func (sq StatQueue) Len() int           { return len(sq) }
func (sq StatQueue) Less(i, j int) bool { return sq[i].Timestamp < sq[j].Timestamp }
func (sq StatQueue) Swap(i, j int) {
	sq[i], sq[j] = sq[j], sq[i]
}

func (sq *StatQueue) Push(x any) {
	*sq = append(*sq, x.(*Bucket))
}

func (sq *StatQueue) Pop() any {
	old := *sq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*sq = old[0 : n-1]

	return item
}
