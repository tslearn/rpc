package main

func main() {
	swapUint32(0x76543210, 5, 5)
}

func swapUint32(v uint32, i int, j int) uint32 {
	i *= 4
	j *= 4
	vi := (v >> i) & 0xF
	v = v&^(0xF<<i) | ((v>>j)&0xF)<<i
	return v&^(0xF<<j) | vi<<j
}

func compareMapItem(m1 mapItem, m2 mapItem) int {
	if m1.fastKey > m2.fastKey {
		return 1
	} else if m1.fastKey < m2.fastKey {
		return -1
	} else {
		if m1.key > m2.key {
			return 1
		} else if m1.key < m2.key {
			return -1
		} else {
			return 0
		}
	}
}

type mapItem struct {
	key     string
	fastKey uint32
	pos     posRecord
}

type posRecord uint64
