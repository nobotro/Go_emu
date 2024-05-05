package register

type Alias uint32

const (
	zero Alias = iota
	ra
	sp
	gp
	tp
	t0
	t1
	t2
	fp
	s1
	a0
	a1
	a2
	a3
	a4
	a5
	a6
	a7
	s2
	s3
	s4
	s5
	s6
	s7
	s8
	s9
	s10
	s11
	t3
	t4
	t5
	t6
)

type RegisterFile struct {
	registers [32]uint32
}

func (reg *RegisterFile) GetRegVal(register uint32) uint32 {

	return reg.registers[register]

}

func (reg *RegisterFile) SetRegVal(register uint32, val uint32) {

	if register > 0 {
		reg.registers[register] = val
	}
}
