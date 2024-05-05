package ram

type Ram struct {
	data [40000]uint32
}

func (ram *Ram) GetLine(address uint32) uint32 {
	return ram.data[address>>2]
}
func (ram *Ram) SetLine(address uint32, data uint32) {
	ram.data[address>>2] = data
}
