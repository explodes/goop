package operations

type opResults []*opResult

func (os opResults) Len() int {
	return len(os)
}

func (os opResults) Less(i, j int) bool {
	return os[i].priority < os[j].priority
}

func (os opResults) Swap(i, j int) {
	os[i], os[j] = os[j], os[i]
}

func (os *opResults) Push(x interface{}) {
	*os = append(*os, x.(*opResult))
}

func (os *opResults) Pop() interface{} {
	old := *os
	n := len(old)
	x := old[n-1]
	*os = old[0 : n-1]
	return x
}