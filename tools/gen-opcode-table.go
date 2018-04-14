package main

import (
	"fmt"
)

// 寻址模式（Addressing Modes）
const (
	_                 byte = iota
	amImmediate            // 1  立即寻址
	amZero                 // 2  零页索引寻址
	amZeroX                // 3  零页直接寻址
	amZeroY                // 4  零页直接寻址
	amAbsolute             // 5  绝对寻址
	amAbsoluteX            // 6  绝对X寻址
	amAbsoluteY            // 7  绝对Y寻址
	amIndirect             // 8  间接寻址
	amIndexedIndirect      // 9  先索引后间接
	amIndirectIndexed      // 10 先间接后索引
	amRelative             // 11 相对寻址
	amImplied              // 12 隐含寻址
	amAccumulator          // 13 累加器寻址
)

type OpCode struct {
	code   byte
	name   string
	size   byte
	cycles byte
	paged  byte
	mode   byte
}

var opcodesTable = [...]OpCode{
	// adc
	{0x69, "ADC", 2, 3, 0, amImmediate},
	{0x65, "ADC", 2, 3, 0, amZero},
	{0x75, "ADC", 2, 4, 0, amZeroX},
	{0x6D, "ADC", 3, 4, 0, amAbsolute},
	{0x7D, "ADC", 3, 4, 1, amAbsoluteX},
	{0x79, "ADC", 3, 4, 1, amAbsoluteY},
	{0x61, "ADC", 2, 6, 0, amIndexedIndirect},
	{0x71, "ADC", 2, 5, 1, amIndirectIndexed},

	// and
	{0x29, "AND", 2, 2, 0, amImmediate},
	{0x25, "AND", 2, 3, 0, amZero},
	{0x35, "AND", 3, 4, 0, amZeroX},
	{0x2D, "AND", 3, 4, 0, amAbsolute},
	{0x3D, "AND", 3, 4, 1, amAbsoluteX},
	{0x39, "AND", 2, 4, 1, amAbsoluteY},
	{0x21, "AND", 2, 6, 0, amIndexedIndirect},
	{0x31, "AND", 2, 5, 1, amIndirectIndexed},

	// asl
	{0x0A, "ASL", 1, 3, 0, amAccumulator},
	{0x06, "ASL", 2, 5, 0, amZero},
	{0x16, "ASL", 2, 6, 0, amZeroX},
	{0x0E, "ASL", 3, 6, 0, amAbsolute},
	{0x1E, "ASL", 3, 7, 0, amAbsoluteX},

	{0x90, "BCC", 2, 2, 0, amRelative},
	{0xB0, "BCS", 2, 2, 0, amRelative},
	{0xF0, "BEQ", 2, 2, 0, amRelative},
	{0x24, "BIT", 2, 3, 0, amZero},
	{0x2C, "BIT", 3, 4, 0, amAbsolute},
	{0x30, "BMI", 2, 2, 0, amRelative},
	{0xD0, "BNE", 2, 2, 0, amRelative},
	{0x10, "BPL", 2, 2, 0, amRelative},

	{0x00, "BRK", 1, 7, 0, amImplied},

	{0x50, "BVC", 2, 2, 0, amRelative},
	{0x70, "BVS", 2, 2, 0, amRelative},

	{0x18, "CLC", 1, 2, 0, amImplied},
	{0xD8, "CLD", 1, 2, 0, amImplied},
	{0x58, "CLI", 1, 2, 0, amImplied},
	{0xB8, "CLV", 1, 2, 0, amImplied},

	// cmp
	{0xC9, "CMP", 2, 2, 0, amImmediate},
	{0xC5, "CMP", 2, 3, 0, amZero},
	{0xD5, "CMP", 2, 4, 0, amZeroX},
	{0xCD, "CMP", 3, 4, 0, amAbsolute},
	{0xDD, "CMP", 3, 4, 1, amAbsoluteX},
	{0xD9, "CMP", 3, 4, 1, amAbsoluteY},
	{0xC1, "CMP", 2, 6, 0, amIndexedIndirect},
	{0xD1, "CMP", 2, 5, 1, amIndirectIndexed},

	{0xE0, "CPX", 2, 2, 0, amImmediate},
	{0xE4, "CPX", 2, 3, 0, amZero},
	{0xEC, "CPX", 3, 4, 0, amAbsolute},

	{0xC0, "CPY", 2, 2, 0, amImmediate},
	{0xC4, "CPY", 2, 3, 0, amZero},
	{0xCC, "CPY", 3, 4, 0, amAbsolute},

	{0xC6, "DEC", 2, 5, 0, amZero},
	{0xD6, "DEC", 2, 6, 0, amZeroX},
	{0xCE, "DEC", 3, 6, 0, amAbsolute},
	{0xDE, "DEC", 3, 7, 0, amAbsoluteX},

	{0xCA, "DEX", 1, 2, 0, amImplied},
	{0x88, "DEY", 1, 2, 0, amImplied},

	{0x49, "EOR", 2, 2, 0, amImmediate},
	{0x45, "EOR", 2, 3, 0, amZero},
	{0x55, "EOR", 2, 4, 0, amZeroX},
	{0x4D, "EOR", 3, 4, 0, amAbsolute},
	{0x5D, "EOR", 3, 4, 1, amAbsoluteX},
	{0x59, "EOR", 3, 4, 1, amAbsoluteY},
	{0x41, "EOR", 2, 6, 0, amIndexedIndirect},
	{0x51, "EOR", 2, 5, 1, amIndirectIndexed},

	{0xE6, "INC", 2, 5, 0, amZero},
	{0xF6, "INC", 2, 6, 0, amZeroX},
	{0xEE, "INC", 3, 6, 0, amAbsolute},
	{0xFE, "INC", 3, 7, 0, amAbsoluteX},

	{0xE8, "INX", 1, 2, 0, amImplied},
	{0xC8, "INY", 1, 2, 0, amImplied},

	{0x4C, "JMP", 3, 3, 0, amAbsolute},
	{0x6C, "JMP", 3, 5, 0, amIndirect},

	{0x20, "JSR", 3, 6, 0, amAbsolute},

	{0xA9, "LDA", 2, 2, 0, amImmediate},
	{0xA5, "LDA", 2, 3, 0, amZero},
	{0xB5, "LDA", 2, 4, 0, amZeroX},
	{0xAD, "LDA", 3, 4, 0, amAbsolute},
	{0xBD, "LDA", 3, 4, 1, amAbsoluteX},
	{0xB9, "LDA", 3, 4, 1, amAbsoluteY},
	{0xA1, "LDA", 2, 6, 0, amIndexedIndirect},
	{0xB1, "LDA", 2, 5, 1, amIndirectIndexed},

	{0xA2, "LDX", 2, 2, 0, amImmediate},
	{0xA6, "LDX", 2, 3, 0, amZero},
	{0xB6, "LDX", 2, 4, 0, amZeroY},
	{0xAE, "LDX", 3, 4, 0, amAbsolute},
	{0xBE, "LDX", 3, 4, 1, amAbsoluteY},

	{0xA0, "LDY", 2, 2, 0, amImmediate},
	{0xA4, "LDY", 2, 3, 0, amZero},
	{0xB4, "LDY", 2, 4, 0, amZeroX},
	{0xAC, "LDY", 3, 4, 0, amAbsolute},
	{0xBC, "LDY", 3, 4, 1, amAbsoluteX},

	{0x4A, "LSR", 1, 2, 0, amAccumulator},
	{0x46, "LSR", 2, 5, 0, amZero},
	{0x56, "LSR", 2, 6, 0, amZeroX},
	{0x4E, "LSR", 3, 6, 0, amAbsolute},
	{0x5E, "LSR", 3, 7, 0, amAbsoluteX},

	{0xEA, "NOP", 1, 2, 0, amImplied},

	{0x09, "ORA", 2, 2, 0, amImmediate},
	{0x05, "ORA", 2, 3, 0, amZero},
	{0x15, "ORA", 2, 4, 0, amZeroX},
	{0x0D, "ORA", 3, 4, 0, amAbsolute},
	{0x1D, "ORA", 3, 4, 1, amAbsoluteX},
	{0x19, "ORA", 3, 4, 1, amAbsoluteY},
	{0x01, "ORA", 2, 6, 0, amIndexedIndirect},
	{0x11, "ORA", 2, 5, 1, amIndirectIndexed},

	{0x48, "PHA", 1, 3, 0, amImplied},
	{0x08, "PHP", 1, 3, 0, amImplied},
	{0x68, "PLA", 1, 4, 0, amImplied},
	{0x28, "PLP", 1, 4, 0, amImplied},

	{0x2A, "ROL", 1, 2, 0, amAccumulator},
	{0x26, "ROL", 2, 5, 0, amZero},
	{0x36, "ROL", 2, 6, 0, amZeroX},
	{0x2E, "ROL", 3, 6, 0, amAbsolute},
	{0x3E, "ROL", 3, 7, 0, amAbsoluteX},

	{0x6A, "ROR", 1, 2, 0, amAccumulator},
	{0x66, "ROR", 2, 5, 0, amZero},
	{0x76, "ROR", 2, 6, 0, amZeroX},
	{0x6E, "ROR", 3, 6, 0, amAbsolute},
	{0x7E, "ROR", 3, 7, 0, amAbsoluteX},

	{0x40, "RTI", 1, 6, 0, amImplied},
	{0x60, "RTS", 1, 6, 0, amImplied},

	{0xE9, "SBC", 2, 2, 0, amImmediate},
	{0xE5, "SBC", 2, 3, 0, amZero},
	{0xF5, "SBC", 2, 4, 0, amZeroX},
	{0xED, "SBC", 3, 4, 0, amAbsolute},
	{0xFD, "SBC", 3, 4, 1, amAbsoluteX},
	{0xF9, "SBC", 3, 4, 1, amAbsoluteY},
	{0xE1, "SBC", 2, 6, 0, amIndexedIndirect},
	{0xF1, "SBC", 2, 5, 1, amIndirectIndexed},

	{0x38, "SEC", 1, 2, 0, amImplied},
	{0xF8, "SED", 1, 2, 0, amImplied},
	{0x78, "SEI", 1, 2, 0, amImplied},

	{0x85, "STA", 3, 5, 0, amZero},
	{0x95, "STA", 3, 5, 0, amZeroX},
	{0x8D, "STA", 3, 5, 0, amAbsolute},
	{0x9D, "STA", 3, 5, 0, amAbsoluteX},
	{0x99, "STA", 3, 5, 0, amAbsoluteY},
	{0x81, "STA", 3, 5, 0, amIndexedIndirect},
	{0x91, "STA", 3, 5, 0, amIndirectIndexed},

	{0x86, "STX", 2, 3, 0, amZero},
	{0x96, "STX", 2, 4, 0, amZeroY},
	{0x8E, "STX", 3, 4, 0, amAbsolute},

	{0x84, "STY", 2, 3, 0, amZero},
	{0x94, "STY", 2, 4, 0, amZeroX},
	{0x8C, "STY", 3, 4, 0, amAbsolute},

	{0xAA, "TAX", 1, 2, 0, amImplied},
	{0xA8, "TAY", 1, 2, 0, amImplied},
	{0xBA, "TSX", 1, 2, 0, amImplied},
	{0x8A, "TXA", 1, 2, 0, amImplied},
	{0x9A, "TXS", 1, 2, 0, amImplied},
	{0x98, "TYA", 1, 2, 0, amImplied},
}

func main() {
	opcTable256 := [256]OpCode{}
	for _, code := range opcodesTable {
		opcTable256[code.code] = code
	}

	modeStr := ""
	sizeStr := ""
	pageStr := ""
	cycleStr := ""

	for i, opc := range opcTable256 {
		modeStr += fmt.Sprintf("%d,", opc.mode)
		sizeStr += fmt.Sprintf("%d,", opc.size)
		pageStr += fmt.Sprintf("%d,", opc.paged)
		cycleStr += fmt.Sprintf("%d,", opc.cycles)

		if (i+1)%16 == 0 {
			modeStr += "\n"
			sizeStr += "\n"
			pageStr += "\n"
			cycleStr += "\n"
		}
	}

	fmt.Printf("modeString:\n%s\n\nsizeStr:\n%s\n\npageStr:\n%s\n\ncycleStr:\n%s\n\n",
		modeStr, sizeStr, pageStr, cycleStr,
	)
}
