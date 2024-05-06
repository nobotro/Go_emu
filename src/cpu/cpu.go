package cpu

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/bits"
	"os"
	"Go_emu/src/ram"
	"Go_emu/src/register"
)

type InstructionType uint32
type MemopsType uint32
type Stage uint32

const (
	R InstructionType = 1
	I InstructionType = 2
	S InstructionType = 3
	B InstructionType = 4
	U InstructionType = 5
	J InstructionType = 6
)

const (
	STORE MemopsType = 1
	LOAD  MemopsType = 2
)

const (
	IF  Stage = 1
	ID  Stage = 2
	IE  Stage = 3
	MEM Stage = 4
	WB  Stage = 5
)

var opcodesMapping = map[uint32]InstructionType{

	0b0110011: R,
	0b0010011: I,
	0b0000011: I,
	0b1100111: I,
	0b0100011: S,
	0b1100011: B,
	0b0110111: U,
	0b1101111: J,
	0b0010111: U,
}

type Memops struct {
	optype    MemopsType
	data      uint32
	address   uint32
	data_mask uint32
	signed    bool
}

type Wbops struct {
	data uint32
	dest uint32
}

type Instruction struct {
	instype   InstructionType
	romline   uint32
	funct7    uint32
	rs2       uint32
	rs1       uint32
	rd        uint32
	funct3    uint32
	opcode    uint32
	imm       uint32
	memop     *Memops
	wbop      *Wbops
	stage     Stage
	rs2_index uint32
	rs1_index uint32
	pc        uint32
}

type Cpu struct {
	regFile     register.RegisterFile //cpu registers
	instStorage [5]*Instruction
	stall       bool   //to stall cpu in case of control and some hazards
	pc          uint32 //program counter
	Ram         ram.Ram
}

func IsBranchIns(inst *Instruction) bool {

	if inst == nil {
		return false
	}

	if inst.instype == B || inst.instype == J ||
		inst.opcode == 1100111 {

		return true
	}
	return false
}

func SignExtend(data uint32, pos uint8) uint32 {
	return uint32(int32(data<<(32-pos)) >> (32 - pos))
}

// includes [end:start]
// end 31<---------------0 start
func SubBits(source uint32, start uint32, end uint32) uint32 {
	var mask = uint32(0xffff_ffff) >> (32 - (end - start) - 1)
	source = (source >> start)
	return source & mask
}

func (inst *Instruction) extractOperands(romline uint32, regFile register.RegisterFile) {

	if inst.instype == R || inst.instype == S || inst.instype == B {

		inst.rs1_index = SubBits(romline, 15, 19)
		inst.rs2_index = SubBits(romline, 20, 24)
		inst.rs1 = regFile.GetRegVal(inst.rs1_index)
		inst.rs2 = regFile.GetRegVal(inst.rs2_index)
		inst.funct3 = SubBits(romline, 12, 14)

	}
	if inst.instype == I {

		inst.rs1_index = SubBits(romline, 15, 19)

		inst.rs1 = regFile.GetRegVal(inst.rs1_index)
		inst.funct3 = SubBits(romline, 12, 14)

	}
	if inst.instype == R || inst.instype == I || inst.instype == U || inst.instype == J {

		inst.rd = SubBits(romline, 7, 11)

	}

	switch inst.instype {
	case R:

		inst.funct7 = SubBits(romline, 25, 31)
	case I:
		inst.imm = SubBits(romline, 20, 31)
	case S:
		inst.imm = SubBits(romline, 25, 31)<<5 | SubBits(romline, 7, 11)

	case B:
		inst.imm = ((SubBits(romline, 31, 31) << 11) |
			(SubBits(romline, 7, 7) << 10) |
			(SubBits(romline, 25, 30) << 4) |
			SubBits(romline, 8, 11)) << 1
	case U:
		inst.imm = SubBits(romline, 12, 31)
	case J:
		inst.imm = ((SubBits(romline, 31, 31) << 19) |
			(SubBits(romline, 12, 19) << 11) |
			(SubBits(romline, 20, 20) << 10) |
			(SubBits(romline, 21, 30))) << 1

	}

}

// stage 1
// get instruction from ram
func (cpu *Cpu) fetchInst(instChannel chan *Instruction) {
	var inst *Instruction = nil
	if !cpu.stall {
		data := cpu.Ram.GetLine(cpu.pc)
		if data != 0 {
			inst = &Instruction{
				romline: data,
				stage:   IF,
				pc:      cpu.pc,
			}
		}
	}

	instChannel <- inst

}

// stage 2
// parse instruction
// get operands
// fetch registers
func (cpu *Cpu) decodeInst(inst *Instruction, instChannelOut chan *Instruction) {

	if inst == nil {
		instChannelOut <- nil
		return
	}

	var opcode = 0b1111111 & inst.romline

	if instype, ok := opcodesMapping[opcode]; ok {

		inst.instype = instype
		inst.opcode = opcode

		inst.extractOperands(inst.romline, cpu.regFile)
		inst.stage = ID
	}

	instChannelOut <- inst

}

// stage 3
// execute opcode logic
func (cpu *Cpu) executeInst(inst *Instruction, instChannelOut chan *Instruction) {
	if inst == nil {
		instChannelOut <- nil
		return
	}
	switch inst.instype {
	case R:
		{
			inst.wbop = &Wbops{

				dest: inst.rd,
			}
			switch inst.funct3 {

			//ADD , SUB
			case 0x0:
				{

					switch inst.funct7 {
					//ADD
					case 0x0:
						inst.wbop.data = inst.rs1 + inst.rs2
					//SUB
					case 0x20:
						inst.wbop.data = inst.rs1 - inst.rs2

					}

				}

			//XOR
			case 0x4:
				inst.wbop.data = inst.rs1 ^ inst.rs2
			//OR
			case 0x6:
				inst.wbop.data = inst.rs1 | inst.rs2

			//SLL
			case 0x1:
				inst.wbop.data = inst.rs1 << (inst.rs2 & 0b11111)
			//SRL , SRA
			case 0x5:
				{

					switch inst.funct7 {
					//SRL
					case 0x0:
						inst.wbop.data = inst.rs1 >> (inst.rs2 & 0b11111)
					//SRA
					case 0x20:
						inst.wbop.data = uint32(int(inst.rs1) >> (inst.rs2 & 0b11111))

					}

				}
			//SLT
			case 0x2:
				var fl uint32 = 0
				if int32(inst.rs1) < int32(inst.rs2) {
					fl = 1
				}
				inst.wbop.data = fl
			//SLTU
			case 0x3:
				var fl uint32 = 0
				if inst.rs1 < inst.rs2 {
					fl = 1
				}
				inst.wbop.data = fl
			}

		}
	case I:

		switch inst.opcode {
		case 0b0010011:
			inst.wbop = &Wbops{

				dest: inst.rd,
			}

			switch inst.funct3 {
			//ADDI
			case 0x0:
				//sign extend
				inst.wbop.data = SignExtend(inst.imm, 12) + inst.rs1
			//XORI
			case 0x4:
				//sign extend
				inst.wbop.data = SignExtend(inst.imm, 12) ^ inst.rs1
			//ORI
			case 0x6:
				//sign extend
				inst.wbop.data = SignExtend(inst.imm, 12) | inst.rs1
			//ANDI
			case 0x7:
				//sign extend
				inst.wbop.data = SignExtend(inst.imm, 12) & inst.rs1
			//SLLI
			case 0x1:

				inst.wbop.data = inst.rs1 << (inst.imm & 0b11111)
			//SRLI,SRAI
			case 0x5:
				shv := SubBits(inst.imm, 0, 4)
				switch SubBits(inst.imm, 5, 11) {
				case 0b0:
					inst.wbop.data = inst.rs1 >> shv
				case 0b0100000:
					inst.wbop.data = uint32(int(inst.rs1) >> shv)
				}
			//SLTI
			case 0x2:
				var fl uint32 = 0
				if int32(inst.rs1) < int32(SignExtend(inst.imm, 12)) {
					fl = 1
				}
				inst.wbop.data = fl
			//SLTIU
			case 0x3:
				var fl uint32 = 0
				if inst.rs1 < SignExtend(inst.imm, 12) {
					fl = 1
				}
				inst.wbop.data = fl

			}

			//MEMORY LOADS
		case 0b0000011:
			inst.memop = &Memops{
				optype: LOAD,
			}
			inst.wbop = &Wbops{

				dest: inst.rd,
			}

			switch inst.funct3 {
			//LB
			case 0x0:
				inst.memop.address = inst.rs1 + SignExtend(inst.imm, 12)
				inst.memop.data_mask = 0xFF
				inst.memop.signed = true
			//LH
			case 0x1:
				inst.memop.address = inst.rs1 + SignExtend(inst.imm, 12)
				inst.memop.data_mask = 0xFFFF
				inst.memop.signed = true
			//LW
			case 0x2:
				inst.memop.address = inst.rs1 + SignExtend(inst.imm, 12)
				inst.memop.data_mask = 0xFFFF_FFFF
				inst.memop.signed = true
			//LBU
			case 0x4:
				inst.memop.address = inst.rs1 + SignExtend(inst.imm, 12)
				inst.memop.data_mask = 0xFF
			//LHU
			case 0x5:
				inst.memop.address = inst.rs1 + SignExtend(inst.imm, 12)
				inst.memop.data_mask = 0xFFFF

			}
			//JALR
		case 0b1100111:

			inst.wbop = &Wbops{

				dest: inst.rd,
				data: inst.pc + 4,
			}
			cpu.pc = inst.rs1 + SignExtend(inst.imm, 12)

		}
	case S:
		inst.memop = &Memops{
			optype: STORE,
		}
		switch inst.funct3 {
		//SB
		case 0x0:
			inst.memop.address = inst.rs1 + SignExtend(inst.imm, 12)
			inst.memop.data_mask = 0xFFFF_FF00
			inst.memop.data = inst.rs2 & 0xFF
		//SH
		case 0x1:
			inst.memop.address = inst.rs1 + SignExtend(inst.imm, 12)
			inst.memop.data_mask = 0xFFFF_0000
			inst.memop.data = inst.rs2 & 0xFFFF
		//SW
		case 0x2:
			inst.memop.address = inst.rs1 + SignExtend(inst.imm, 12)
			inst.memop.data_mask = 0x0
			inst.memop.data = inst.rs2 & 0xFFFF_FFFF

		}
	case B:

		switch inst.funct3 {
		//BEQ
		case 0x0:
			if inst.rs1 == inst.rs2 {
				cpu.pc = inst.pc + SignExtend(inst.imm, 12)
			}
		//BNE
		case 0x1:
			if inst.rs1 != inst.rs2 {
				cpu.pc = inst.pc + SignExtend(inst.imm, 12)
			}
		//BLT
		case 0x4:
			if int32(inst.rs1) < int32(inst.rs2) {
				cpu.pc = inst.pc + SignExtend(inst.imm, 12)
			}
		//BGE
		case 0x5:
			if int32(inst.rs1) >= int32(inst.rs2) {
				cpu.pc = inst.pc + SignExtend(inst.imm, 12)
			}
		//BLTU
		case 0x6:
			if inst.rs1 < inst.rs2 {
				cpu.pc = inst.pc + SignExtend(inst.imm, 12)
			}
		//BGEU
		case 0x7:
			if inst.rs1 >= inst.rs2 {
				cpu.pc = inst.pc + SignExtend(inst.imm, 12)
			}
		}
	case J:
		//JAL
		inst.wbop = &Wbops{

			dest: inst.rd,
			data: inst.pc + 4,
		}
		cpu.pc = inst.pc + SignExtend(inst.imm, 20)
	case U:
		switch inst.opcode {
		//LUI
		case 0b0110111:
			inst.wbop = &Wbops{

				dest: inst.rd,
				data: SignExtend(inst.imm<<12, 31),
			}
		//AUIPC
		case 0b0010111:
			inst.wbop = &Wbops{

				dest: inst.rd,
				data: inst.pc + SignExtend(inst.imm<<12, 31),
			}
		}
	}
	inst.stage = IE
	instChannelOut <- inst

}

// stage 4
// load and store data in memory if necessary
func (cpu *Cpu) memOps(inst *Instruction, instChannelOut chan *Instruction) {
	if inst == nil {
		instChannelOut <- nil
		return
	}
	if inst.memop != nil {

		if inst.memop.optype == LOAD {

			data := cpu.Ram.GetLine(inst.memop.address) & inst.memop.data_mask
			if inst.memop.signed {
				data = SignExtend(data, uint8(bits.OnesCount32(inst.memop.data_mask)))
			}
			inst.wbop.data = data

		} else {

			data := cpu.Ram.GetLine(inst.memop.address)
			data = (data & inst.memop.data_mask) | inst.memop.data
			cpu.Ram.SetLine(inst.memop.address, data)

		}
	}
	inst.stage = MEM
	instChannelOut <- inst

}

// stage 5
// store result into registers if necessary
func (cpu *Cpu) writeBack(inst *Instruction, instChannelOut chan *Instruction) {
	if inst == nil {
		instChannelOut <- nil
		return
	}

	if inst.wbop != nil {

		cpu.regFile.SetRegVal(inst.wbop.dest, inst.wbop.data)
	}
	inst.stage = WB

	instChannelOut <- inst

}

func (cpu *Cpu) hazardHandler() {

	////if branch instruction already executed,remove stall
	if cpu.stall && IsBranchIns(cpu.instStorage[4]) {
		cpu.stall = false
	}

	////if decoded instruction is branch inst,
	////stall cpu(not fetch new instuction until cur inst executed)
	if IsBranchIns(cpu.instStorage[1]) {

		cpu.instStorage[0] = nil
		cpu.stall = true
	}

	if cpu.instStorage[1] != nil && (cpu.instStorage[1].rs2_index != 0 || cpu.instStorage[1].rs1_index != 0) {

		rs1_found := false
		rs2_found := false

		for i := 2; i < len(cpu.instStorage); i++ {

			if cpu.instStorage[i] == nil || cpu.instStorage[i].wbop == nil {
				continue
			}

			//detect interlock for load instructions
			//if load instruction

			if i == 2 && cpu.instStorage[2].opcode == 0b0000011 &&
				(cpu.instStorage[2].wbop.dest == cpu.instStorage[1].rs1_index ||
					cpu.instStorage[2].wbop.dest == cpu.instStorage[1].rs2_index) {

				//insert nop in pipeline

				cpu.instStorage[0] = cpu.instStorage[1]
				cpu.pc = cpu.instStorage[1].pc
				cpu.instStorage[1] = nil
				break

			}
			if !rs1_found && cpu.instStorage[i].wbop.dest == cpu.instStorage[1].rs1_index {

				cpu.instStorage[1].rs1 = cpu.instStorage[i].wbop.data
				rs1_found = true
			}
			if !rs2_found && cpu.instStorage[i].wbop.dest == cpu.instStorage[1].rs2_index {

				cpu.instStorage[1].rs2 = cpu.instStorage[i].wbop.data
				rs2_found = true
			}

			if rs2_found && rs1_found {
				break
			}

		}

	}
}

func (cpu *Cpu) ClockCycle() {

	fetched := make(chan *Instruction)
	decoded := make(chan *Instruction)
	executed := make(chan *Instruction)
	memopsed := make(chan *Instruction)
	wbed := make(chan *Instruction)

	go cpu.fetchInst(fetched)
	go cpu.decodeInst(cpu.instStorage[0], decoded)
	go cpu.executeInst(cpu.instStorage[1], executed)
	go cpu.memOps(cpu.instStorage[2], memopsed)
	go cpu.writeBack(cpu.instStorage[3], wbed)

	cpu.instStorage[4] = <-wbed
	cpu.instStorage[3] = <-memopsed
	cpu.instStorage[2] = <-executed
	cpu.instStorage[1] = <-decoded
	cpu.instStorage[0] = <-fetched

	cpu.hazardHandler()
	//if we don't have branch instruction
	if cpu.instStorage[0] != nil && cpu.instStorage[0].pc == cpu.pc && !cpu.stall {
		cpu.pc = cpu.pc + 4
	}

}

func (cpu *Cpu) LoadRom(path string) {
	file, error := os.Open(path)
	if error != nil {
		log.Fatal(error)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buffer := make([]byte, 4)
	i := uint32(0)
	for {
		_, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		romLine := binary.LittleEndian.Uint32(buffer)
		cpu.Ram.SetLine(i, romLine)
		i += 4

	}
}
