package persist

type BranchClocks []*BranchClock

func (b *BlockRef) Size() uint64 {
	return b.Upper - b.Lower
}
