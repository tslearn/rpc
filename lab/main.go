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
