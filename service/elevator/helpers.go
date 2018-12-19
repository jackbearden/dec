package elevator

const (
	mask0, bit0 = (1 << (1 << iota)) - 1, 1 << iota
	mask1, bit1
	mask2, bit2
	mask3, bit3
)

func MSB16(x uint16) (out int) {
	if x == 0 {
		return -1
	}
	if x&^mask3 != 0 {
		x >>= bit3
		out |= bit3
	}
	if x&^mask2 != 0 {
		x >>= bit2
		out |= bit2
	}
	if x&^mask1 != 0 {
		x >>= bit1
		out |= bit1
	}
	if x&^mask0 != 0 {
		out |= bit0
	}
	return
}

func LSB16(x uint16) (out int) {
	if x == 0 {
		return -1
	}
	if x&mask3 == 0 {
		x >>= bit3
		out |= bit3
	}
	if x&mask2 == 0 {
		x >>= bit2
		out |= bit2
	}
	if x&mask1 == 0 {
		x >>= bit1
		out |= bit1
	}
	if x&mask0 == 0 {
		out |= bit0
	}
	return
}
