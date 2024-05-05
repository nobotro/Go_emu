package cpu

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"risc-v_emulator/src/register"
	"testing"
)

func TestSubBits(t *testing.T) {

	var input_results = []uint32{
		SubBits(0b1000_1111_1111_1100_1111_0000_1111_1111, 0, 7),
		SubBits(0b0000_0000_1111_1100_1111_0000_1111_1111, 0, 31),
		SubBits(0b0000_0000_0000_1100_1111_0000_1111_1111, 1, 4),
		SubBits(0b1111_1111_1111_1111_1111_0001_1111_1111, 8, 8),
		SubBits(0b1000_1000_1111_1100_1111_0000_1111_1111, 7, 20),
		SubBits(0b0000_1111_1100_0111_1111_1111_1001_1001, 8, 13),
		SubBits(0b1100_1111_1100_0111_1111_1100_0000_1001, 1, 30),
	}

	var expected_results = []uint32{

		0b1111_1111,
		0b0000_0000_1111_1100_1111_0000_1111_1111,
		0b1_111,
		0b1,
		0b1_1100_1111_0000_1,
		0b11_1111,
		0b100_1111_1100_0111_1111_1100_0000_100,
	}
	for i, v := range input_results {
		if v != expected_results[i] {
			t.Errorf("\"TestSubBits()\" FAILED, expected -> %32b, got -> %32b", expected_results[i], v)
			return
		}
	}

}

func TestSignExtend(t *testing.T) {

	var input_results = []uint32{
		SignExtend(0b1000_1000_1111_1100_1111_0000_1111_1111, 32),
		SignExtend(0b0000_0000_1111_1100_1111_0000_1111_1111, 32),
		SignExtend(0b0000_0000_0000_1100_1111_0000_1111_1111, 16),
		SignExtend(0b1000_1000_1111_1100_1111_0000_1111_1111, 16),
		SignExtend(0b1000_1000_1111_1100_1111_0000_1111_1111, 20),
		SignExtend(0b0000_1111_1100_00111_1111_1111_1001_1001, 8),
		SignExtend(0b0000_1111_1100_00111_1111_1100_0000_1001, 8),
	}

	var expected_results = []uint32{

		0b1000_1000_1111_1100_1111_0000_1111_1111,
		0b0000_0000_1111_1100_1111_0000_1111_1111,
		0b1111_1111_1111_1111_1111_0000_1111_1111,
		0b1111_1111_1111_1111_1111_0000_1111_1111,
		0b1111_1111_1111_1100_1111_0000_1111_1111,
		0b1111_1111_1111_1111_1111_1111_1001_1001,
		0b0000_0000_0000_0000_0000_0000_0000_1001,
	}
	for i, v := range input_results {
		if v != expected_results[i] {
			t.Errorf("\"TestSignExtend()\" FAILED, expected -> %32b, got -> %32b", expected_results[i], v)
			return
		}
	}

}

func TestDecodeInst(t *testing.T) {

	//init registers with test data
	regFile := register.RegisterFile{}
	for i := 0; i < 32; i++ {
		regFile.SetRegVal(uint32(i), uint32(i))
	}

	outchan := make(chan *Instruction)

	cpu := Cpu{
		regFile: regFile,
	}

	var input_instructions = []*Instruction{

		//sub x1, x2, x3
		&Instruction{
			romline:   0b0100_0000_0011_0001_0000_0000_1011_0011,
			funct7:    0b010_0000,
			rs2:       0b0001_1,
			rs1:       0b0001_0,
			rs1_index: 0b0001_0,
			rs2_index: 0b0001_1,
			funct3:    0,
			rd:        0b00001,
			opcode:    0b011001_1,
			instype:   R,
			stage:     ID,
		},
		//xori x30, x20, 60
		&Instruction{
			romline:   0b0000_0011_1100_1010_0100_1111_0001_0011,
			imm:       0b0000_0011_1100,
			rs1:       0b1010_0,
			rs1_index: 0b1010_0,
			funct3:    0b100,
			rd:        0b1111_0,
			opcode:    0b001_0011,
			instype:   I,
			stage:     ID,
		},
		//lbu x1, 45(x8)
		&Instruction{
			romline:   0b0000_0010_1101_0100_0100_0000_1000_0011,
			imm:       0b0000_0010_1101,
			rs1:       0b0100_0,
			rs1_index: 0b0100_0,
			funct3:    0b100,
			rd:        0b0000_1,
			opcode:    0b000_0011,
			instype:   I,
			stage:     ID,
		},
		//sh x1, 200(x23)
		&Instruction{
			romline:   0b0000_1100_0001_1011_1001_0100_0010_0011,
			imm:       0b0000_1100_1000,
			rs2:       0b0_0001,
			rs1:       0b1011_1,
			rs1_index: 0b1011_1,
			rs2_index: 0b0_0001,
			funct3:    0b001,
			opcode:    0b010_0011,
			instype:   S,
			stage:     ID,
		},

		//bge x1, x2, 44
		&Instruction{
			romline:   0b0000_0010_0010_0000_1101_0110_0110_0011,
			imm:       0b000_0001_0110_0,
			rs2:       0b0_0010,
			rs1:       0b0000_1,
			rs1_index: 0b0000_1,
			rs2_index: 0b0_0010,
			funct3:    0b101,
			opcode:    0b110_0011,
			instype:   B,
			stage:     ID,
		},

		//jal x1, 64
		&Instruction{
			romline: 0b0000_0100_0000_0000_0000_0000_1110_1111,
			imm:     0b000_0000_0000_0010_00000,
			rd:      0b0000_1,
			opcode:  0b110_1111,
			instype: J,
			stage:   ID,
		},
		//jal x1, 10940
		&Instruction{
			romline: 0b0010_1011_1101_0000_0010_0000_1110_1111,
			imm:     0b000_0001_0101_0101_1110_0,
			rd:      0b0000_1,
			opcode:  0b110_1111,
			instype: J,
			stage:   ID,
		},
		//auipc x11, 72
		&Instruction{
			romline: 0b0000_0000_0000_0100_1000_0101_1001_0111,
			imm:     0b0000_0000_0000_0100_1000,
			rd:      0b0101_1,
			opcode:  0b001_0111,
			instype: U,
			stage:   ID,
		},
	}
	for _, v := range input_instructions {

		inst := &Instruction{romline: v.romline}
		go cpu.decodeInst(inst, outchan)
		decoded := <-outchan
		if !reflect.DeepEqual(v, decoded) {
			t.Errorf("\"TestDecodeInst()\" FAILED, expected -> %v, got -> %v opcode -> %07b romline->  %032b", decoded, v, v.opcode, v.romline)
			return
		}
	}

}

// Note that all control input instruction has same pc after execution as imm
// (IsBranchIns(inst) && cpu.pc != decoded.imm)
func TestExecuteInst(t *testing.T) {

	regFile := register.RegisterFile{}
	outchan := make(chan *Instruction)
	for i := 0; i < 32; i++ {
		regFile.SetRegVal(uint32(i), uint32(i))
	}

	var input_instructions = []*Instruction{

		//sub x1, x2, x3
		&Instruction{
			romline: 0b0100_0000_0011_0001_0000_0000_1011_0011,
			wbop: &Wbops{
				dest: 1,
				data: 0xffffffff, //-1
			},
		},
		//xori x30, x20, 60
		&Instruction{
			romline: 0b0000_0011_1100_1010_0100_1111_0001_0011,
			wbop: &Wbops{
				dest: 30,
				data: 0x00000028, //40
			},
		},
		//lbu x1, 45(x8)
		&Instruction{
			romline: 0b0000_0010_1101_0100_0100_0000_1000_0011,
			memop: &Memops{
				optype:    LOAD,
				address:   0x35, //53,
				data_mask: 0xFF,
			},
		},
		//sh x1, 200(x23)
		&Instruction{
			romline: 0b0000_1100_0001_1011_1001_0100_0010_0011,
			memop: &Memops{
				optype:    STORE,
				address:   0xdf, //223,
				data_mask: 0xFFFF_0000,
				data:      1,
			},
		},
		//bge x2, x1, 44
		&Instruction{
			romline: 0b0000_0010_0001_0001_0101_0110_0110_0011,
			//cpu pc must be 48 after execution this
		},
		//jal x1, 64
		&Instruction{
			romline: 0b0000_0100_0000_0000_0000_0000_1110_1111,
			wbop: &Wbops{
				dest: 1,
				data: 4,
			},
			//cpu pc must be 64 after execution this

		},
		//jal x1, 10940
		&Instruction{
			romline: 0b0010_1011_1101_0000_0010_0000_1110_1111,
			wbop: &Wbops{
				dest: 1,
				data: 4,
			},
			//cpu pc must be 10940 after execution this
		},
		//auipc x11, 72
		&Instruction{
			romline: 0b0000_0000_0000_0100_1000_0101_1001_0111,
			wbop: &Wbops{
				dest: 11,
				data: 0x48000,
			},
		},
	}
	for _, v := range input_instructions {
		inst := &Instruction{romline: v.romline}
		cpu := Cpu{
			regFile: regFile,
		}
		go cpu.decodeInst(inst, outchan)
		decoded := <-outchan
		go cpu.executeInst(decoded, outchan)
		executed := <-outchan
		if v.wbop != nil && !reflect.DeepEqual(v.wbop, executed.wbop) ||
			(v.memop != nil && !reflect.DeepEqual(v.memop, executed.memop)) ||
			(IsBranchIns(inst) && cpu.pc != decoded.imm) {
			t.Errorf("\"TestExecuteInst()\" FAILED, expected -> %v, got -> %v opcode -> %07b romline->  %032b", executed, v, v.opcode, v.romline)
			return
		}
	}

}

func TestDataHazardHandler(t *testing.T) {

	//test data hazard forwarding
	//1:add x1, x2, x3  //5
	//2:add x4, x2, x5   //7
	//3:add x6, x1, x7   //12
	//4:add x9, x1, x6   //17
	//5:add x10, x6, x6   //24

	var possible_data_hazard = []uint32{
		0b0000_0000_0011_0001_0000_0000_1011_0011,
		0b0000_0000_0101_0001_0000_0010_0011_0011,
		0b0000_0000_0111_0000_1000_0011_0011_0011,
		0b0000_0000_0110_0000_1000_0100_1011_0011,
		0b0000_0000_0110_0011_0000_0101_0011_0011,
	}
	regFile := register.RegisterFile{}

	for i := 0; i < 32; i++ {
		regFile.SetRegVal(uint32(i), uint32(i))
	}
	cpu := Cpu{
		regFile: regFile,
	}
	for i, v := range possible_data_hazard {
		cpu.Ram.SetLine(uint32(i*4), v)

	}
	for i := 0; i < 10000; i++ {
		cpu.ClockCycle()
	}

	if cpu.regFile.GetRegVal(1) != 5 ||
		cpu.regFile.GetRegVal(2) != 2 ||
		cpu.regFile.GetRegVal(3) != 3 ||
		cpu.regFile.GetRegVal(4) != 7 ||
		cpu.regFile.GetRegVal(5) != 5 ||
		cpu.regFile.GetRegVal(6) != 12 ||
		cpu.regFile.GetRegVal(7) != 7 ||
		cpu.regFile.GetRegVal(9) != 17 ||
		cpu.regFile.GetRegVal(10) != 24 {
		t.Errorf("\"TestDataHazardHandler()\" FAILED")
		return

	}

}

func TestInterlockDataHazardHandler(t *testing.T) {

	//test interlock data hazard
	//sw x10,0x50(x0)
	//add x1, x2, x3
	//add x4, x2, x5
	//add x6, x1, x7
	//add x9, x1, x6
	//lw x1, 0x50(x0)
	//add x3, x1, x1
	//add x10, x6, x6

	var possible_data_hazard = []uint32{

		0b0000_0100_1010_0000_0010_1000_0010_0011,
		0b0000_0000_0011_0001_0000_0000_1011_0011,
		0b0000_0000_0101_0001_0000_0010_0011_0011,
		0b0000_0000_0111_0000_1000_0011_0011_0011,
		0b0000_0000_0110_0000_1000_0100_1011_0011,
		0b0000_0101_0000_0000_0010_0000_1000_0011,
		0b0000_0000_0001_0000_1000_0001_1011_0011,
		0b0000_0000_0110_0011_0000_0101_0011_0011,
	}
	regFile := register.RegisterFile{}

	for i := 0; i < 32; i++ {
		regFile.SetRegVal(uint32(i), uint32(i))
	}
	cpu := Cpu{
		regFile: regFile,
	}
	for i, v := range possible_data_hazard {
		cpu.Ram.SetLine(uint32(i*4), v)

	}
	for i := 0; i < 10000; i++ {
		cpu.ClockCycle()
	}

	if cpu.regFile.GetRegVal(1) != 10 ||
		cpu.regFile.GetRegVal(2) != 2 ||
		cpu.regFile.GetRegVal(3) != 20 ||
		cpu.regFile.GetRegVal(4) != 7 ||
		cpu.regFile.GetRegVal(5) != 5 ||
		cpu.regFile.GetRegVal(6) != 12 ||
		cpu.regFile.GetRegVal(7) != 7 ||
		cpu.regFile.GetRegVal(9) != 17 ||
		cpu.regFile.GetRegVal(10) != 24 {
		t.Errorf("\"TestInterlockDataHazardHandler()\" FAILED")
		return

	}

}

func TestControlHazardHandler(t *testing.T) {

	//test interlock data hazard
	// sw x10,0x50(x0)
	// addi x1, x0,100
	// addi x2, x0,200
	// add x3 ,x0 ,x0
	// add x8 ,x0 ,x0
	// blt x1, x2, 12
	// addi x3, x3, 15
	// addi x3, x3, 45
	// addi x3, x3, 65
	// lw x1, 0x50(x0)
	// add x10, x6, x6
	// jal x1,12
	// addi x8, x8, 15
	// addi x8, x8, 45
	// addi x8, x8, 65
	// addi x7, x3, 40

	var possible_data_hazard = []uint32{

		0b0000_0100_1010_0000_0010_1000_0010_0011,
		0b0000_0110_0100_0000_0000_0000_1001_0011,
		0b0000_1100_1000_0000_0000_0001_0001_0011,
		0b0000_0000_0000_0000_0000_0001_1011_0011,
		0b0000_0000_0000_0000_0000_0100_0011_0011,
		0b0000_0000_0010_0000_1100_0110_0110_0011,
		0b0000_0000_1111_0001_1000_0001_1001_0011,
		0b0000_0010_1101_0001_1000_0001_1001_0011,
		0b0000_0100_0001_0001_1000_0001_1001_0011,
		0b0000_0101_0000_0000_0010_0000_1000_0011,
		0b0000_0000_0110_0011_0000_0101_0011_0011,
		0b0000_0000_1100_0000_0000_0000_1110_1111,
		0b0000_0000_1111_0100_0000_0100_0001_0011,
		0b0000_0010_1101_0100_0000_0100_0001_0011,
		0b0000_0100_0001_0100_0000_0100_0001_0011,
		0b0000_0010_1000_0001_1000_0011_1001_0011,
	}
	regFile := register.RegisterFile{}

	for i := 0; i < 32; i++ {
		regFile.SetRegVal(uint32(i), uint32(i))
	}
	cpu := Cpu{
		regFile: regFile,
	}
	for i, v := range possible_data_hazard {
		cpu.Ram.SetLine(uint32(i*4), v)

	}
	for i := 0; i < 10000; i++ {
		cpu.ClockCycle()
	}

	if cpu.regFile.GetRegVal(1) != 48 ||
		cpu.regFile.GetRegVal(2) != 200 ||
		cpu.regFile.GetRegVal(3) != 65 ||
		cpu.regFile.GetRegVal(4) != 4 ||
		cpu.regFile.GetRegVal(5) != 5 ||
		cpu.regFile.GetRegVal(6) != 6 ||
		cpu.regFile.GetRegVal(7) != 105 ||
		cpu.regFile.GetRegVal(8) != 65 ||
		cpu.regFile.GetRegVal(9) != 9 ||
		cpu.regFile.GetRegVal(10) != 12 {
		t.Errorf("\"TestControlHazardHandler()\" FAILED")
		return

	}

}

func TestRomExecution(t *testing.T) {
	file, error := os.Open("/home/user/Risc-v-chip8/sdk/rom")
	if error != nil {
		log.Fatal(error)
	}

	// close the file at the end of the program

	cpu := Cpu{}
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
		// print the image in hexadecimal format
		romline := binary.LittleEndian.Uint32(buffer)

		cpu.Ram.SetLine(uint32(i*4), romline)
		i += 1

	}
	for i := 0; i < 1000000; i++ {
		cpu.ClockCycle()
	}
	if !(cpu.Ram.GetLine(30028) == 1000 &&
		cpu.Ram.GetLine(30032) == 120 &&
		cpu.Ram.GetLine(30036) == 120 &&
		cpu.Ram.GetLine(30040) == 120 &&
		cpu.Ram.GetLine(30044) == 120 &&
		cpu.Ram.GetLine(30048) == 0) {
		t.Errorf("\"TestRomExecution()\" FAILED")

	}
}
